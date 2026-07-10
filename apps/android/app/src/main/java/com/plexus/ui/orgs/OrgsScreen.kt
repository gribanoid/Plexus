package com.plexus.ui.orgs

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Logout
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.plexus.data.api.PlexusApi
import com.plexus.data.models.Organization
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

data class OrgsUiState(
    val isLoading: Boolean = false,
    val orgs: List<Organization> = emptyList(),
    val error: String? = null,
    val isCreating: Boolean = false,
    val createError: String? = null,
)

@HiltViewModel
class OrgsViewModel @Inject constructor(private val api: PlexusApi) : ViewModel() {
    private val _uiState = MutableStateFlow(OrgsUiState(isLoading = true))
    val uiState: StateFlow<OrgsUiState> = _uiState.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching { api.listOrgs() }
                .onSuccess { res -> _uiState.update { it.copy(isLoading = false, orgs = res.items) } }
                .onFailure { e -> _uiState.update { it.copy(isLoading = false, error = e.message) } }
        }
    }

    fun createOrg(name: String, slug: String?, onSuccess: () -> Unit) {
        viewModelScope.launch {
            _uiState.update { it.copy(isCreating = true, createError = null) }
            runCatching {
                val body = mutableMapOf("name" to name)
                if (!slug.isNullOrBlank()) body["slug"] = slug
                api.createOrg(body)
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
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun OrgsScreen(
    onOrgSelected: (slug: String) -> Unit,
    onNotifications: () -> Unit,
    onLogout: () -> Unit,
    viewModel: OrgsViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
    var showCreateDialog by remember { mutableStateOf(false) }

    if (showCreateDialog) {
        CreateOrgDialog(
            onDismiss = {
                showCreateDialog = false
                viewModel.clearCreateError()
            },
            onConfirm = { name, slug ->
                viewModel.createOrg(name, slug) {
                    showCreateDialog = false
                    viewModel.clearCreateError()
                }
            },
            isSubmitting = uiState.isCreating,
            error = uiState.createError,
        )
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Workspaces") },
                actions = {
                    IconButton(onClick = onNotifications) {
                        Icon(Icons.Default.Notifications, contentDescription = "Notifications")
                    }
                    IconButton(onClick = onLogout) {
                        Icon(Icons.Default.Logout, contentDescription = "Sign out")
                    }
                },
            )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = { showCreateDialog = true }) {
                Icon(Icons.Default.Add, contentDescription = "New workspace")
            }
        },
    ) { padding ->
        when {
            uiState.isLoading -> Box(
                Modifier.fillMaxSize().padding(padding),
                contentAlignment = Alignment.Center,
            ) { CircularProgressIndicator() }

            uiState.error != null -> Box(
                Modifier.fillMaxSize().padding(padding),
                contentAlignment = Alignment.Center,
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(uiState.error!!, color = MaterialTheme.colorScheme.error)
                    Spacer(Modifier.height(8.dp))
                    Button(onClick = { viewModel.load() }) { Text("Retry") }
                }
            }

            uiState.orgs.isEmpty() -> Box(
                Modifier.fillMaxSize().padding(padding),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    "No workspaces yet.\nCreate your first one.",
                    style = MaterialTheme.typography.bodyLarge,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }

            else -> LazyColumn(
                modifier = Modifier.fillMaxSize().padding(padding),
                contentPadding = PaddingValues(vertical = 8.dp),
            ) {
                items(uiState.orgs, key = { it.id }) { org ->
                    OrgItem(org = org, onClick = { onOrgSelected(org.slug) })
                    HorizontalDivider()
                }
            }
        }
    }
}

@Composable
private fun OrgItem(org: Organization, onClick: () -> Unit) {
    ListItem(
        modifier = Modifier.clickable(onClick = onClick),
        headlineContent = { Text(org.name, fontWeight = FontWeight.Medium) },
        supportingContent = {
            Text(
                "${org.plan.replaceFirstChar { it.uppercase() }} · ${org.myRole ?: "member"}",
                style = MaterialTheme.typography.bodySmall,
            )
        },
        leadingContent = {
            Surface(
                shape = MaterialTheme.shapes.small,
                color = MaterialTheme.colorScheme.primaryContainer,
                modifier = Modifier.size(40.dp),
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text(
                        org.name.first().uppercaseChar().toString(),
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.onPrimaryContainer,
                    )
                }
            }
        },
    )
}
