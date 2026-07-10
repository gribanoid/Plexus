package com.plexus.ui.projects

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.plexus.data.api.PlexusApi
import com.plexus.data.models.Project
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ProjectsUiState(
    val isLoading: Boolean = false,
    val orgName: String? = null,
    val projects: List<Project> = emptyList(),
    val error: String? = null,
    val isCreating: Boolean = false,
    val createError: String? = null,
)

@HiltViewModel
class ProjectsViewModel @Inject constructor(
    private val api: PlexusApi,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val orgSlug: String = checkNotNull(savedStateHandle["orgSlug"])
    private val _uiState = MutableStateFlow(ProjectsUiState(isLoading = true))
    val uiState: StateFlow<ProjectsUiState> = _uiState.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                val org = api.getOrg(orgSlug)
                val projects = api.listProjects(orgSlug)
                org to projects
            }.onSuccess { (org, res) ->
                _uiState.update {
                    it.copy(isLoading = false, orgName = org.name, projects = res.items)
                }
            }.onFailure { e ->
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun createProject(
        name: String,
        key: String?,
        description: String?,
        onSuccess: (projectKey: String) -> Unit,
    ) {
        viewModelScope.launch {
            _uiState.update { it.copy(isCreating = true, createError = null) }
            runCatching {
                val body = mutableMapOf("name" to name)
                if (!key.isNullOrBlank()) body["key"] = key
                if (!description.isNullOrBlank()) body["description"] = description
                api.createProject(orgSlug, body)
            }.onSuccess { project ->
                load()
                _uiState.update { it.copy(isCreating = false) }
                onSuccess(project.key)
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
fun ProjectsScreen(
    orgSlug: String,
    onProjectSelected: (projectKey: String) -> Unit,
    onBack: () -> Unit,
    onBacklogSelected: ((projectKey: String) -> Unit)? = null,
    viewModel: ProjectsViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
    var showCreateDialog by remember { mutableStateOf(false) }

    if (showCreateDialog) {
        CreateProjectDialog(
            orgName = uiState.orgName,
            onDismiss = {
                showCreateDialog = false
                viewModel.clearCreateError()
            },
            onConfirm = { name, key, description ->
                viewModel.createProject(name, key, description) { _ ->
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
                title = { Text(uiState.orgName ?: orgSlug) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
            )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = { showCreateDialog = true }) {
                Icon(Icons.Default.Add, contentDescription = "New project")
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

            uiState.projects.isEmpty() -> Box(
                Modifier.fillMaxSize().padding(padding),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    "No projects yet.\nCreate your first project.",
                    style = MaterialTheme.typography.bodyLarge,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
            }

            else -> LazyColumn(
                modifier = Modifier.fillMaxSize().padding(padding),
                contentPadding = PaddingValues(vertical = 8.dp),
            ) {
                items(uiState.projects, key = { it.id }) { project ->
                    ProjectItem(
                        project = project,
                        onBoardClick = { onProjectSelected(project.key) },
                        onBacklogClick = { onBacklogSelected?.invoke(project.key) },
                    )
                    HorizontalDivider()
                }
            }
        }
    }
}

@Composable
private fun ProjectItem(
    project: Project,
    onBoardClick: () -> Unit,
    onBacklogClick: () -> Unit,
) {
    ListItem(
        modifier = Modifier.clickable(onClick = onBoardClick),
        headlineContent = { Text(project.name, fontWeight = FontWeight.Medium) },
        supportingContent = {
            Text(
                project.description ?: "No description",
                style = MaterialTheme.typography.bodySmall,
                maxLines = 1,
            )
        },
        leadingContent = {
            Surface(
                shape = MaterialTheme.shapes.small,
                color = MaterialTheme.colorScheme.secondaryContainer,
                modifier = Modifier.size(40.dp),
            ) {
                Box(contentAlignment = Alignment.Center) {
                    Text(
                        project.key,
                        style = MaterialTheme.typography.labelMedium,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colorScheme.onSecondaryContainer,
                    )
                }
            }
        },
        trailingContent = {
            TextButton(onClick = onBacklogClick) { Text("Backlog") }
        },
    )
}
