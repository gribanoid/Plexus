package com.plexus.ui.backlog

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ArrowUpward
import androidx.compose.material.icons.filled.Remove
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.plexus.data.api.PlexusApi
import com.plexus.data.models.Issue
import com.plexus.data.models.IssueType
import com.plexus.data.models.Sprint
import com.plexus.data.models.Status
import com.plexus.ui.issues.CreateIssueDialog
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

data class BacklogUiState(
    val isLoading: Boolean = false,
    val sprints: List<Sprint> = emptyList(),
    val issues: List<Issue> = emptyList(),
    val statuses: List<Status> = emptyList(),
    val issueTypes: List<IssueType> = emptyList(),
    val isLoadingIssueTypes: Boolean = false,
    val isCreating: Boolean = false,
    val createError: String? = null,
    val error: String? = null,
    val actingSprintId: String? = null,
)

@HiltViewModel
class BacklogViewModel @Inject constructor(
    private val api: PlexusApi,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val orgSlug: String = checkNotNull(savedStateHandle["orgSlug"])
    private val projectKey: String = checkNotNull(savedStateHandle["projectKey"])

    private val _uiState = MutableStateFlow(BacklogUiState(isLoading = true))
    val uiState: StateFlow<BacklogUiState> = _uiState.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                val sprintsDef = async { api.listSprints(orgSlug, projectKey) }
                val issuesDef = async { api.listIssues(orgSlug, projectKey) }
                val statusesDef = async { api.listStatuses(orgSlug, projectKey) }
                Triple(sprintsDef.await(), issuesDef.await(), statusesDef.await())
            }.onSuccess { (sp, is_, st) ->
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        sprints = sp.items,
                        issues = is_.items,
                        statuses = st.items,
                    )
                }
            }.onFailure { e ->
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun statusFor(issue: Issue) = _uiState.value.statuses.firstOrNull { it.id == issue.statusId }

    fun loadIssueTypes() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoadingIssueTypes = true, createError = null) }
            runCatching { api.listIssueTypes(orgSlug, projectKey) }
                .onSuccess { res ->
                    _uiState.update { it.copy(isLoadingIssueTypes = false, issueTypes = res.items) }
                }
                .onFailure {
                    _uiState.update { it.copy(isLoadingIssueTypes = false) }
                }
        }
    }

    fun createIssue(title: String, typeId: String, priority: String, onSuccess: () -> Unit) {
        viewModelScope.launch {
            _uiState.update { it.copy(isCreating = true, createError = null) }
            runCatching {
                api.createIssue(
                    orgSlug,
                    projectKey,
                    mapOf(
                        "title" to title,
                        "type_id" to typeId,
                        "priority" to priority,
                    ),
                )
            }.onSuccess {
                load()
                _uiState.update { it.copy(isCreating = false) }
                onSuccess()
            }.onFailure { e ->
                _uiState.update { it.copy(isCreating = false, createError = e.message) }
            }
        }
    }

    fun clearCreateError() {
        _uiState.update { it.copy(createError = null) }
    }

    fun startSprint(sprintId: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(actingSprintId = sprintId) }
            runCatching { api.startSprint(orgSlug, projectKey, sprintId) }
                .onSuccess { load() }
            _uiState.update { it.copy(actingSprintId = null) }
        }
    }

    fun completeSprint(sprintId: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(actingSprintId = sprintId) }
            runCatching { api.completeSprint(orgSlug, projectKey, sprintId) }
                .onSuccess { load() }
            _uiState.update { it.copy(actingSprintId = null) }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BacklogScreen(
    projectKey: String,
    onBack: () -> Unit = {},
    onIssueClick: ((issueNumber: Int) -> Unit)? = null,
    viewModel: BacklogViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
    var expandedSprints by remember { mutableStateOf(setOf<String>()) }
    var showCreateDialog by remember { mutableStateOf(false) }

    LaunchedEffect(showCreateDialog) {
        if (showCreateDialog) viewModel.loadIssueTypes()
    }

    if (showCreateDialog) {
        CreateIssueDialog(
            onDismiss = {
                showCreateDialog = false
                viewModel.clearCreateError()
            },
            onConfirm = { title, typeId, priority ->
                viewModel.createIssue(title, typeId, priority) {
                    showCreateDialog = false
                    viewModel.clearCreateError()
                }
            },
            issueTypes = uiState.issueTypes,
            isLoadingTypes = uiState.isLoadingIssueTypes,
            isSubmitting = uiState.isCreating,
            error = uiState.createError,
        )
    }

    fun toggle(id: String) {
        expandedSprints = if (expandedSprints.contains(id)) expandedSprints - id else expandedSprints + id
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("$projectKey — Backlog") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
            )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = { showCreateDialog = true }) {
                Icon(Icons.Default.Add, contentDescription = "New issue")
            }
        },
    ) { padding ->
        if (uiState.isLoading) {
            Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
            return@Scaffold
        }

        LazyColumn(
            modifier = Modifier.fillMaxSize().padding(padding),
            contentPadding = PaddingValues(bottom = 24.dp),
        ) {
            // Sprints
            uiState.sprints.forEach { sprint ->
                val sprintIssues = uiState.issues.filter { it.sprintId == sprint.id }
                val isExpanded = expandedSprints.contains(sprint.id)
                val sprintActionLabel = when (sprint.state) {
                    "active" -> "Complete sprint"
                    "future" -> "Start sprint"
                    else -> null
                }

                item(key = "header_${sprint.id}") {
                    SprintHeader(
                        sprint = sprint,
                        issueCount = sprintIssues.size,
                        isExpanded = isExpanded,
                        onToggle = { toggle(sprint.id) },
                        actionLabel = sprintActionLabel,
                        isActionLoading = uiState.actingSprintId == sprint.id,
                        onAction = when (sprint.state) {
                            "active" -> ({ viewModel.completeSprint(sprint.id) })
                            "future" -> ({ viewModel.startSprint(sprint.id) })
                            else -> null
                        },
                    )
                    HorizontalDivider()
                }

                if (isExpanded) {
                    items(sprintIssues, key = { it.id }) { issue ->
                        IssueListItem(
                            issue = issue,
                            statusName = viewModel.statusFor(issue)?.name,
                            projectKey = projectKey,
                            onClick = onIssueClick?.let { cb -> { cb(issue.number) } },
                        )
                        HorizontalDivider(modifier = Modifier.padding(start = 56.dp))
                    }
                }
            }

            // Backlog (no sprint)
            val backlog = uiState.issues.filter { it.sprintId == null }
            val isBacklogExpanded = expandedSprints.contains("backlog")

            item(key = "header_backlog") {
                SprintHeader(
                    sprint = Sprint(id = "backlog", name = "Backlog", goal = null,
                        state = "future", startDate = null, endDate = null),
                    issueCount = backlog.size,
                    isExpanded = isBacklogExpanded,
                    onToggle = { toggle("backlog") },
                )
                HorizontalDivider()
            }

            if (isBacklogExpanded) {
                items(backlog, key = { it.id }) { issue ->
                    IssueListItem(
                        issue = issue,
                        statusName = viewModel.statusFor(issue)?.name,
                        projectKey = projectKey,
                        onClick = onIssueClick?.let { cb -> { cb(issue.number) } },
                    )
                    HorizontalDivider(modifier = Modifier.padding(start = 56.dp))
                }
            }
        }
    }
}

@Composable
private fun SprintHeader(
    sprint: Sprint,
    issueCount: Int,
    isExpanded: Boolean,
    onToggle: () -> Unit,
    actionLabel: String? = null,
    isActionLoading: Boolean = false,
    onAction: (() -> Unit)? = null,
) {
    ListItem(
        modifier = Modifier,
        headlineContent = {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Text(sprint.name, style = MaterialTheme.typography.titleSmall)
                if (sprint.state == "active") {
                    Surface(
                        shape = MaterialTheme.shapes.small,
                        color = MaterialTheme.colorScheme.tertiaryContainer,
                    ) {
                        Text(
                            "ACTIVE",
                            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onTertiaryContainer,
                        )
                    }
                }
            }
        },
        supportingContent = {
            Text("$issueCount issues", style = MaterialTheme.typography.bodySmall)
        },
        trailingContent = {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                if (actionLabel != null && onAction != null) {
                    TextButton(
                        onClick = onAction,
                        enabled = !isActionLoading,
                    ) {
                        if (isActionLoading) {
                            CircularProgressIndicator(Modifier.size(16.dp), strokeWidth = 2.dp)
                        } else {
                            Text(actionLabel)
                        }
                    }
                }
                TextButton(onClick = onToggle) {
                    Text(if (isExpanded) "Collapse" else "Expand")
                }
            }
        },
    )
}

@Composable
private fun IssueListItem(
    issue: Issue,
    statusName: String?,
    projectKey: String,
    onClick: (() -> Unit)? = null,
) {
    ListItem(
        modifier = if (onClick != null) Modifier.clickable(onClick = onClick) else Modifier,
        leadingContent = {
            val (icon, tint) = when (issue.priority) {
                "urgent", "high" -> Icons.Default.ArrowUpward to MaterialTheme.colorScheme.error
                else -> Icons.Default.Remove to MaterialTheme.colorScheme.outline
            }
            Icon(icon, contentDescription = issue.priority, tint = tint, modifier = Modifier.size(20.dp))
        },
        headlineContent = {
            Text(issue.title, maxLines = 2, overflow = TextOverflow.Ellipsis)
        },
        supportingContent = {
            Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                Text(
                    "$projectKey-${issue.number}",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.outline,
                )
                statusName?.let {
                    Text(it, style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
            }
        },
        trailingContent = {
            issue.storyPoints?.let { sp ->
                Surface(
                    shape = MaterialTheme.shapes.small,
                    color = MaterialTheme.colorScheme.secondaryContainer,
                ) {
                    Text(
                        "${sp.toInt()}",
                        modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
                        style = MaterialTheme.typography.labelSmall,
                    )
                }
            }
        },
    )
}
