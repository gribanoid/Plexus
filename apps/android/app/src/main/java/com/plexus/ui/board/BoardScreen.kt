package com.plexus.ui.board

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ArrowUpward
import androidx.compose.material.icons.filled.Remove
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.plexus.data.models.Issue
import com.plexus.data.models.Status
import com.plexus.ui.issues.CreateIssueDialog

@Composable
fun BoardScreen(
    projectKey: String,
    onBack: () -> Unit = {},
    onIssueClick: ((issueNumber: Int) -> Unit)? = null,
    viewModel: BoardViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
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

    Scaffold(topBar = {
        TopAppBar(
            title = { Text("$projectKey — Board") },
            navigationIcon = {
                IconButton(onClick = onBack) {
                    Icon(
                        imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                        contentDescription = "Back",
                    )
                }
            },
        )
    }, floatingActionButton = {
        FloatingActionButton(onClick = { showCreateDialog = true }) {
            Icon(Icons.Default.Add, contentDescription = "New issue")
        }
    }) { padding ->
        if (uiState.isLoading) {
            Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else {
            Row(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding)
                    .horizontalScroll(rememberScrollState())
                    .padding(horizontal = 12.dp, vertical = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                uiState.statuses.forEach { status ->
                    BoardColumn(
                        status = status,
                        issues = uiState.issuesFor(status.id),
                        allStatuses = uiState.statuses,
                        projectKey = projectKey,
                        onIssueClick = onIssueClick,
                        onMoveIssue = { issueNumber, newStatusId ->
                            viewModel.moveIssue(issueNumber, newStatusId)
                        },
                    )
                }
            }
        }
    }
}

@Composable
fun BoardColumn(
    status: Status,
    issues: List<Issue>,
    allStatuses: List<Status>,
    projectKey: String,
    onIssueClick: ((issueNumber: Int) -> Unit)? = null,
    onMoveIssue: ((issueNumber: Int, newStatusId: String) -> Unit)? = null,
) {
    Column(
        modifier = Modifier
            .width(250.dp)
            .fillMaxHeight(),
    ) {
        // Column header
        Row(
            modifier = Modifier.padding(bottom = 8.dp, start = 4.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            Surface(
                shape = MaterialTheme.shapes.small,
                color = parseHexColor(status.color),
                modifier = Modifier.size(8.dp),
            ) {}
            Text(
                status.name,
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            Spacer(Modifier.weight(1f))
            Text(
                "${issues.size}",
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.outline,
            )
        }

        Surface(
            modifier = Modifier.fillMaxHeight(),
            shape = MaterialTheme.shapes.medium,
            color = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f),
        ) {
            LazyColumn(
                contentPadding = PaddingValues(8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
                modifier = Modifier.fillMaxHeight(),
            ) {
                items(issues, key = { it.id }) { issue ->
                    IssueCard(
                        issue = issue,
                        projectKey = projectKey,
                        statuses = allStatuses,
                        onClick = onIssueClick?.let { cb -> { cb(issue.number) } },
                        onMoveToStatus = onMoveIssue?.let { move ->
                            { newStatusId -> move(issue.number, newStatusId) }
                        },
                    )
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class, ExperimentalFoundationApi::class)
@Composable
fun IssueCard(
    issue: Issue,
    projectKey: String,
    statuses: List<Status> = emptyList(),
    onClick: (() -> Unit)? = null,
    onMoveToStatus: ((newStatusId: String) -> Unit)? = null,
) {
    var showMoveMenu by remember { mutableStateOf(false) }

    Box {
        Card(
            modifier = Modifier.fillMaxWidth().then(
                if (onClick != null || onMoveToStatus != null) {
                    Modifier.combinedClickable(
                        onClick = { onClick?.invoke() },
                        onLongClick = {
                            if (onMoveToStatus != null && statuses.isNotEmpty()) {
                                showMoveMenu = true
                            }
                        },
                    )
                } else {
                    Modifier
                }
            ),
            elevation = CardDefaults.cardElevation(defaultElevation = 1.dp),
        ) {
            Column(modifier = Modifier.padding(10.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                Text(
                    issue.title,
                    style = MaterialTheme.typography.bodyMedium,
                    maxLines = 3,
                    overflow = TextOverflow.Ellipsis,
                )
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    PriorityChip(issue.priority)
                    Text(
                        "$projectKey-${issue.number}",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.outline,
                    )
                    Spacer(Modifier.weight(1f))
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
                }
            }
        }

        DropdownMenu(
            expanded = showMoveMenu,
            onDismissRequest = { showMoveMenu = false },
        ) {
            statuses.filter { it.id != issue.statusId }.forEach { target ->
                DropdownMenuItem(
                    text = { Text("Move to ${target.name}") },
                    onClick = {
                        showMoveMenu = false
                        onMoveToStatus?.invoke(target.id)
                    },
                )
            }
        }
    }
}

@Composable
fun PriorityChip(priority: String) {
    val (color, icon) = when (priority) {
        "urgent" -> MaterialTheme.colorScheme.error to Icons.Default.ArrowUpward
        "high" -> Color(0xFFF97316) to Icons.Default.ArrowUpward
        else -> MaterialTheme.colorScheme.outline to Icons.Default.Remove
    }
    Icon(icon, contentDescription = priority, tint = color, modifier = Modifier.size(14.dp))
}

fun parseHexColor(hex: String): Color {
    return try {
        val clean = hex.removePrefix("#")
        val long = clean.toLong(16)
        Color(
            red = ((long shr 16) and 0xFF) / 255f,
            green = ((long shr 8) and 0xFF) / 255f,
            blue = (long and 0xFF) / 255f,
        )
    } catch (e: Exception) {
        Color.Gray
    }
}
