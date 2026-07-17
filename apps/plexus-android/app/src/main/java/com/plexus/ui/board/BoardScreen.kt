package com.plexus.ui.board

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.MoreVert
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.plexus.data.models.Issue
import com.plexus.data.models.IssueType
import com.plexus.data.models.OrgMember
import com.plexus.data.models.Status
import com.plexus.ui.issues.CreateIssueDialog

@Composable
fun BoardScreen(
    projectKey: String,
    onBack: () -> Unit = {},
    onIssueClick: ((issueNumber: Int) -> Unit)? = null,
    embedded: Boolean = false,
    viewModel: BoardViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
    var showCreateDialog by remember { mutableStateOf(false) }
    var createStatusId by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(showCreateDialog) {
        if (showCreateDialog) viewModel.loadIssueTypes()
    }

    if (showCreateDialog) {
        CreateIssueDialog(
            onDismiss = {
                showCreateDialog = false
                createStatusId = null
                viewModel.clearCreateError()
            },
            onConfirm = { title, typeId, priority ->
                viewModel.createIssue(title, typeId, priority, createStatusId) {
                    showCreateDialog = false
                    createStatusId = null
                    viewModel.clearCreateError()
                }
            },
            issueTypes = uiState.issueTypes,
            isLoadingTypes = uiState.isLoadingIssueTypes,
            isSubmitting = uiState.isCreating,
            error = uiState.createError,
        )
    }

    val content: @Composable (PaddingValues) -> Unit = { padding ->
        when {
            uiState.isLoading -> Box(
                modifier = Modifier.fillMaxSize().padding(padding),
                contentAlignment = Alignment.Center,
            ) { CircularProgressIndicator() }

            uiState.error != null -> Box(
                modifier = Modifier.fillMaxSize().padding(padding),
                contentAlignment = Alignment.Center,
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(uiState.error!!, color = MaterialTheme.colorScheme.error)
                    Spacer(Modifier.height(8.dp))
                    Button(onClick = { viewModel.load() }) { Text("Retry") }
                }
            }

            else -> Row(
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
                        issueTypes = uiState.issueTypes,
                        members = uiState.members,
                        onIssueClick = onIssueClick,
                        onMoveIssue = { issueNumber, newStatusId ->
                            viewModel.moveIssue(issueNumber, newStatusId)
                        },
                        onCreate = {
                            createStatusId = status.id
                            showCreateDialog = true
                        },
                    )
                }
            }
        }
    }

    if (embedded) {
        Box(Modifier.fillMaxSize()) {
            content(PaddingValues(0.dp))
        }
    } else {
        Scaffold(
            topBar = {
                TopAppBar(
                    title = { Text("$projectKey — Board") },
                    navigationIcon = {
                        IconButton(onClick = onBack) {
                            Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                        }
                    },
                )
            },
            floatingActionButton = {
                FloatingActionButton(onClick = {
                    createStatusId = null
                    showCreateDialog = true
                }) {
                    Icon(Icons.Default.Add, contentDescription = "New issue")
                }
            },
            content = content,
        )
    }
}

@Composable
fun BoardColumn(
    status: Status,
    issues: List<Issue>,
    allStatuses: List<Status>,
    projectKey: String,
    issueTypes: List<IssueType>,
    members: List<OrgMember>,
    onIssueClick: ((issueNumber: Int) -> Unit)? = null,
    onMoveIssue: ((issueNumber: Int, newStatusId: String) -> Unit)? = null,
    onCreate: (() -> Unit)? = null,
) {
    Column(
        modifier = Modifier
            .width(280.dp)
            .fillMaxHeight(),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 12.dp, vertical = 10.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Text(
                status.name.uppercase(),
                style = MaterialTheme.typography.labelMedium,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
            Text(
                "${issues.size}",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.outline,
            )
            Spacer(Modifier.weight(1f))
            Icon(
                Icons.Default.MoreVert,
                contentDescription = null,
                modifier = Modifier.size(18.dp),
                tint = MaterialTheme.colorScheme.outline,
            )
        }

        Surface(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth(),
            shape = RoundedCornerShape(8.dp),
            color = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f),
        ) {
            LazyColumn(
                contentPadding = PaddingValues(10.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
                modifier = Modifier.fillMaxHeight(),
            ) {
                items(issues, key = { it.id }) { issue ->
                    IssueCard(
                        issue = issue,
                        projectKey = projectKey,
                        issueType = issueTypes.firstOrNull { it.id == issue.typeId },
                        assignee = members.firstOrNull { it.id == issue.assigneeId },
                        statuses = allStatuses,
                        onClick = onIssueClick?.let { cb -> { cb(issue.number) } },
                        onMoveToStatus = onMoveIssue?.let { move ->
                            { newStatusId -> move(issue.number, newStatusId) }
                        },
                    )
                }

                if (onCreate != null) {
                    item {
                        TextButton(
                            onClick = onCreate,
                            modifier = Modifier.fillMaxWidth(),
                        ) {
                            Icon(Icons.Default.Add, contentDescription = null, modifier = Modifier.size(16.dp))
                            Spacer(Modifier.width(4.dp))
                            Text("Create")
                        }
                    }
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
    issueType: IssueType? = null,
    assignee: OrgMember? = null,
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
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface),
            elevation = CardDefaults.cardElevation(defaultElevation = 1.dp),
            shape = RoundedCornerShape(8.dp),
        ) {
            Column(
                modifier = Modifier.padding(12.dp),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                Text(
                    issue.title,
                    style = MaterialTheme.typography.bodyMedium,
                    maxLines = 3,
                    overflow = TextOverflow.Ellipsis,
                    color = MaterialTheme.colorScheme.onSurface,
                )
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    Box(
                        modifier = Modifier
                            .size(14.dp)
                            .clip(RoundedCornerShape(3.dp))
                            .background(parseHexColor(issueType?.color ?: "#0052CC")),
                    )
                    Text(
                        "$projectKey-${issue.number}",
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                    )
                    Spacer(Modifier.weight(1f))
                    if (assignee != null) {
                        Box(
                            modifier = Modifier
                                .size(22.dp)
                                .clip(CircleShape)
                                .background(MaterialTheme.colorScheme.primaryContainer),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(
                                assignee.displayName.take(1).uppercase(),
                                style = MaterialTheme.typography.labelSmall,
                                fontWeight = FontWeight.Bold,
                                color = MaterialTheme.colorScheme.onPrimaryContainer,
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

fun parseHexColor(hex: String): Color {
    return try {
        val clean = hex.removePrefix("#")
        val long = clean.toLong(16)
        Color(
            red = ((long shr 16) and 0xFF) / 255f,
            green = ((long shr 8) and 0xFF) / 255f,
            blue = (long and 0xFF) / 255f,
        )
    } catch (_: Exception) {
        Color.Gray
    }
}
