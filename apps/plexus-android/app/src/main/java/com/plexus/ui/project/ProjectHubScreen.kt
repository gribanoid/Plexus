package com.plexus.ui.project

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.FilterList
import androidx.compose.material.icons.filled.MoreVert
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.plexus.data.api.PlexusApi
import com.plexus.data.models.CreateStatusBody
import com.plexus.data.models.Status
import com.plexus.ui.backlog.BacklogScreen
import com.plexus.ui.board.BoardScreen
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

enum class ProjectTab(val label: String) {
    Board("Board"),
    Backlog("Backlog"),
    Roadmap("Roadmap"),
    Reports("Reports"),
    Settings("Settings"),
}

private val statusCategories = listOf(
    "todo" to "To Do",
    "in_progress" to "In Progress",
    "done" to "Done",
)

private val colorPresets = listOf(
    "#4C9AFF", "#6B7280", "#22C55E", "#F97316", "#EF4444", "#8B5CF6", "#EAB308",
)

data class ProjectHubUiState(
    val projectName: String? = null,
    val statuses: List<Status> = emptyList(),
    val isLoading: Boolean = false,
    val isSaving: Boolean = false,
    val error: String? = null,
)

@HiltViewModel
class ProjectHubViewModel @Inject constructor(
    private val api: PlexusApi,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val orgSlug: String = checkNotNull(savedStateHandle["orgSlug"])
    private val projectKey: String = checkNotNull(savedStateHandle["projectKey"])

    private val _uiState = MutableStateFlow(ProjectHubUiState(isLoading = true))
    val uiState: StateFlow<ProjectHubUiState> = _uiState.asStateFlow()

    init {
        load()
    }

    fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                val projects = api.listProjects(orgSlug).items
                val project = projects.firstOrNull { it.key == projectKey }
                val statuses = api.listStatuses(orgSlug, projectKey).items.sortedBy { it.position }
                project?.name to statuses
            }.onSuccess { (name, statuses) ->
                _uiState.update {
                    it.copy(isLoading = false, projectName = name ?: projectKey, statuses = statuses)
                }
            }.onFailure { e ->
                _uiState.update {
                    it.copy(isLoading = false, projectName = projectKey, error = e.message)
                }
            }
        }
    }

    fun createStatus(name: String, color: String, category: String, onDone: () -> Unit = {}) {
        viewModelScope.launch {
            _uiState.update { it.copy(isSaving = true, error = null) }
            runCatching {
                api.createStatus(orgSlug, projectKey, CreateStatusBody(name, color, category))
            }.onSuccess {
                _uiState.update { it.copy(isSaving = false) }
                load()
                onDone()
            }.onFailure { e ->
                _uiState.update { it.copy(isSaving = false, error = e.message) }
            }
        }
    }

    fun deleteStatus(statusId: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isSaving = true, error = null) }
            runCatching {
                api.deleteStatus(orgSlug, projectKey, statusId)
            }.onSuccess {
                _uiState.update { it.copy(isSaving = false) }
                load()
            }.onFailure { e ->
                _uiState.update { it.copy(isSaving = false, error = e.message) }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProjectHubScreen(
    orgSlug: String,
    projectKey: String,
    onBack: () -> Unit,
    onIssueClick: (issueNumber: Int) -> Unit,
    viewModel: ProjectHubViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
    var selectedTab by remember { mutableIntStateOf(0) }
    val tabs = ProjectTab.entries

    // Keep keys in composition for nested ViewModels that read SavedStateHandle.
    @Suppress("UNUSED_VARIABLE")
    val routeKeys = orgSlug to projectKey

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Column {
                        Text(
                            uiState.projectName ?: projectKey,
                            fontWeight = FontWeight.SemiBold,
                            style = MaterialTheme.typography.titleMedium,
                        )
                        Text(
                            projectKey,
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }
                },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    IconButton(onClick = { /* filter stub */ }) {
                        Icon(Icons.Default.FilterList, contentDescription = "Filter")
                    }
                    IconButton(onClick = { /* menu stub */ }) {
                        Icon(Icons.Default.MoreVert, contentDescription = "More")
                    }
                },
            )
        },
    ) { padding ->
        Column(modifier = Modifier.fillMaxSize().padding(padding)) {
            ScrollableTabRow(
                selectedTabIndex = selectedTab,
                edgePadding = 12.dp,
            ) {
                tabs.forEachIndexed { index, tab ->
                    Tab(
                        selected = selectedTab == index,
                        onClick = { selectedTab = index },
                        text = {
                            Text(
                                tab.label,
                                fontWeight = if (selectedTab == index) FontWeight.SemiBold else FontWeight.Normal,
                            )
                        },
                    )
                }
            }

            when (tabs[selectedTab]) {
                ProjectTab.Board -> BoardScreen(
                    projectKey = projectKey,
                    embedded = true,
                    onIssueClick = onIssueClick,
                )
                ProjectTab.Backlog -> BacklogScreen(
                    projectKey = projectKey,
                    embedded = true,
                    onIssueClick = onIssueClick,
                )
                ProjectTab.Roadmap -> ComingSoonPane("Roadmap")
                ProjectTab.Reports -> ComingSoonPane("Reports")
                ProjectTab.Settings -> ProjectSettingsPane(
                    projectKey = projectKey,
                    statuses = uiState.statuses,
                    isSaving = uiState.isSaving,
                    error = uiState.error,
                    onCreate = { name, color, category ->
                        viewModel.createStatus(name, color, category)
                    },
                    onDelete = { viewModel.deleteStatus(it) },
                )
            }
        }
    }
}

@Composable
private fun ComingSoonPane(feature: String) {
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            Text(feature, style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.SemiBold)
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                "Coming soon",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun ProjectSettingsPane(
    projectKey: String,
    statuses: List<Status>,
    isSaving: Boolean,
    error: String?,
    onCreate: (name: String, color: String, category: String) -> Unit,
    onDelete: (statusId: String) -> Unit,
) {
    var name by remember { mutableStateOf("") }
    var color by remember { mutableStateOf("#4C9AFF") }
    var category by remember { mutableStateOf("todo") }
    var categoryExpanded by remember { mutableStateOf(false) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(rememberScrollState())
            .padding(16.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Text("Project settings", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
        Text(
            "Manage workflow columns for $projectKey. Other settings are available on the web.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )

        if (error != null) {
            Text(error, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
        }

        Text("Workflow", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)

        OutlinedTextField(
            value = name,
            onValueChange = { name = it },
            label = { Text("Name") },
            placeholder = { Text("In Review") },
            singleLine = true,
            modifier = Modifier.fillMaxWidth(),
        )

        Text("Color", style = MaterialTheme.typography.labelMedium)
        Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
            colorPresets.forEach { hex ->
                val parsed = parseHexColor(hex)
                Box(
                    modifier = Modifier
                        .size(28.dp)
                        .clip(CircleShape)
                        .background(parsed)
                        .then(
                            if (color == hex) {
                                Modifier.border(2.dp, MaterialTheme.colorScheme.onSurface, CircleShape)
                            } else {
                                Modifier
                            },
                        )
                        .clickable { color = hex },
                )
            }
        }

        ExposedDropdownMenuBox(
            expanded = categoryExpanded,
            onExpandedChange = { categoryExpanded = it },
        ) {
            OutlinedTextField(
                value = statusCategories.firstOrNull { it.first == category }?.second ?: category,
                onValueChange = {},
                readOnly = true,
                label = { Text("Category") },
                trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = categoryExpanded) },
                modifier = Modifier
                    .menuAnchor()
                    .fillMaxWidth(),
            )
            ExposedDropdownMenu(
                expanded = categoryExpanded,
                onDismissRequest = { categoryExpanded = false },
            ) {
                statusCategories.forEach { (value, label) ->
                    DropdownMenuItem(
                        text = { Text(label) },
                        onClick = {
                            category = value
                            categoryExpanded = false
                        },
                    )
                }
            }
        }

        Button(
            onClick = {
                val trimmed = name.trim()
                if (trimmed.isNotEmpty()) {
                    onCreate(trimmed, color, category)
                    name = ""
                }
            },
            enabled = !isSaving && name.trim().isNotEmpty(),
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text(if (isSaving) "Adding…" else "Add status")
        }

        HorizontalDivider()

        if (statuses.isEmpty()) {
            Text("No statuses yet", color = MaterialTheme.colorScheme.outline)
        } else {
            statuses.forEach { status ->
                ListItem(
                    headlineContent = { Text(status.name) },
                    supportingContent = {
                        Text(
                            status.category.replace('_', ' '),
                            style = MaterialTheme.typography.bodySmall,
                        )
                    },
                    leadingContent = {
                        Box(
                            modifier = Modifier
                                .size(10.dp)
                                .clip(CircleShape)
                                .background(parseHexColor(status.color)),
                        )
                    },
                    trailingContent = {
                        IconButton(
                            onClick = { onDelete(status.id) },
                            enabled = !isSaving,
                        ) {
                            Icon(
                                Icons.Default.Delete,
                                contentDescription = "Delete status",
                                tint = MaterialTheme.colorScheme.error,
                            )
                        }
                    },
                )
            }
        }
    }
}

private fun parseHexColor(hex: String): Color {
    val cleaned = hex.removePrefix("#")
    val value = cleaned.toLongOrNull(16) ?: return Color.Gray
    val r = ((value shr 16) and 0xFF) / 255f
    val g = ((value shr 8) and 0xFF) / 255f
    val b = (value and 0xFF) / 255f
    return Color(r, g, b)
}
