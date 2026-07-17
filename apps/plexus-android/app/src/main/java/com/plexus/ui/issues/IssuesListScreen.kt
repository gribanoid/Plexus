package com.plexus.ui.issues

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.plexus.data.api.PlexusApi
import com.plexus.data.models.Issue
import com.plexus.data.models.Organization
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

data class IssueListItem(
    val issue: Issue,
    val orgSlug: String,
    val projectKey: String,
    val projectName: String,
)

data class IssuesListUiState(
    val isLoading: Boolean = false,
    val orgs: List<Organization> = emptyList(),
    val selectedOrgSlug: String? = null,
    val items: List<IssueListItem> = emptyList(),
    val error: String? = null,
)

@HiltViewModel
class IssuesListViewModel @Inject constructor(
    private val api: PlexusApi,
) : ViewModel() {
    private val _uiState = MutableStateFlow(IssuesListUiState(isLoading = true))
    val uiState: StateFlow<IssuesListUiState> = _uiState.asStateFlow()

    init {
        loadOrgsAndIssues()
    }

    fun loadOrgsAndIssues(orgSlug: String? = null) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                val orgs = api.listOrgs().items
                val slug = orgSlug ?: _uiState.value.selectedOrgSlug ?: orgs.firstOrNull()?.slug
                if (slug == null) {
                    return@runCatching Triple(orgs, null as String?, emptyList<IssueListItem>())
                }
                val projects = api.listProjects(slug).items
                val results = projects.map { project ->
                    async {
                        runCatching {
                            api.listIssues(slug, project.key).items.map { issue ->
                                IssueListItem(issue, slug, project.key, project.name)
                            }
                        }
                    }
                }.awaitAll()
                val items = results.mapNotNull { it.getOrNull() }.flatten()
                    .sortedByDescending { it.issue.updatedAt }
                val firstError = results.firstOrNull { it.isFailure }?.exceptionOrNull()?.message
                if (items.isEmpty() && firstError != null) {
                    throw IllegalStateException(firstError)
                }
                Triple(orgs, slug, items)
            }.onSuccess { (orgs, slug, items) ->
                _uiState.update {
                    it.copy(isLoading = false, orgs = orgs, selectedOrgSlug = slug, items = items)
                }
            }.onFailure { e ->
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun selectOrg(slug: String) {
        loadOrgsAndIssues(slug)
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun IssuesListScreen(
    onIssueClick: (orgSlug: String, projectKey: String, issueNumber: Int) -> Unit,
    viewModel: IssuesListViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
    var orgExpanded by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(title = { Text("Issues") })
        },
    ) { padding ->
        Column(Modifier = Modifier.fillMaxSize().padding(padding)) {
            if (uiState.orgs.isNotEmpty()) {
                val selected = uiState.orgs.firstOrNull { it.slug == uiState.selectedOrgSlug }
                ExposedDropdownMenuBox(
                    expanded = orgExpanded,
                    onExpandedChange = { orgExpanded = it },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp, vertical = 8.dp),
                ) {
                    OutlinedTextField(
                        value = selected?.name ?: "Select workspace",
                        onValueChange = {},
                        readOnly = true,
                        label = { Text("Workspace") },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = orgExpanded) },
                        modifier = Modifier
                            .menuAnchor()
                            .fillMaxWidth(),
                    )
                    ExposedDropdownMenu(
                        expanded = orgExpanded,
                        onDismissRequest = { orgExpanded = false },
                    ) {
                        uiState.orgs.forEach { org ->
                            DropdownMenuItem(
                                text = { Text(org.name) },
                                onClick = {
                                    orgExpanded = false
                                    viewModel.selectOrg(org.slug)
                                },
                            )
                        }
                    }
                }
            }

            when {
                uiState.isLoading -> Box(
                    Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) { CircularProgressIndicator() }

                uiState.error != null -> Box(
                    Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(uiState.error!!, color = MaterialTheme.colorScheme.error)
                        Spacer(Modifier.height(8.dp))
                        Button(onClick = { viewModel.loadOrgsAndIssues() }) { Text("Retry") }
                    }
                }

                uiState.items.isEmpty() -> Box(
                    Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        "No issues yet",
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                }

                else -> LazyColumn(contentPadding = PaddingValues(vertical = 4.dp)) {
                    items(uiState.items, key = { "${it.projectKey}-${it.issue.id}" }) { item ->
                        ListItem(
                            modifier = Modifier.clickable {
                                onIssueClick(item.orgSlug, item.projectKey, item.issue.number)
                            },
                            headlineContent = {
                                Text(
                                    item.issue.title,
                                    fontWeight = FontWeight.Medium,
                                    maxLines = 2,
                                    overflow = TextOverflow.Ellipsis,
                                )
                            },
                            supportingContent = {
                                Text(
                                    "${item.projectKey}-${item.issue.number} · ${item.projectName}",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                                )
                            },
                        )
                        HorizontalDivider()
                    }
                }
            }
        }
    }
}
