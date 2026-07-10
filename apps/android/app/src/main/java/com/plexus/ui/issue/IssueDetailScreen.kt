package com.plexus.ui.issue

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.Send
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
import com.plexus.data.models.Comment
import com.plexus.data.models.Issue
import com.plexus.data.models.OrgMember
import com.plexus.data.models.Status
import com.plexus.ui.issues.EditIssueDialog
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

data class IssueDetailUiState(
    val isLoading: Boolean = false,
    val issue: Issue? = null,
    val statuses: List<Status> = emptyList(),
    val members: List<OrgMember> = emptyList(),
    val comments: List<Comment> = emptyList(),
    val error: String? = null,
    val isSendingComment: Boolean = false,
    val isSaving: Boolean = false,
    val saveError: String? = null,
)

@HiltViewModel
class IssueDetailViewModel @Inject constructor(
    private val api: PlexusApi,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val orgSlug: String = checkNotNull(savedStateHandle["orgSlug"])
    private val projectKey: String = checkNotNull(savedStateHandle["projectKey"])
    private val issueNumber: Int = checkNotNull(savedStateHandle.get<String>("issueNumber")?.toIntOrNull())

    private val _uiState = MutableStateFlow(IssueDetailUiState(isLoading = true))
    val uiState: StateFlow<IssueDetailUiState> = _uiState.asStateFlow()

    init { load() }

    fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                val issueDef = async { api.getIssue(orgSlug, projectKey, issueNumber) }
                val commentsDef = async { api.listComments(orgSlug, projectKey, issueNumber) }
                val statusesDef = async { api.listStatuses(orgSlug, projectKey) }
                val membersDef = async { api.listOrgMembers(orgSlug) }
                val issue = issueDef.await()
                val comments = commentsDef.await()
                val statuses = statusesDef.await()
                val members = membersDef.await()
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        issue = issue,
                        comments = comments.items,
                        statuses = statuses.items,
                        members = members.items,
                    )
                }
            }.onFailure { e ->
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun sendComment(body: String) {
        if (body.isBlank()) return
        viewModelScope.launch {
            _uiState.update { it.copy(isSendingComment = true) }
            runCatching {
                api.createComment(orgSlug, projectKey, issueNumber, mapOf("body" to body))
                // Reload comments
                val updated = api.listComments(orgSlug, projectKey, issueNumber)
                _uiState.update { it.copy(comments = updated.items) }
            }.onFailure { /* silently ignore */ }
            _uiState.update { it.copy(isSendingComment = false) }
        }
    }

    fun statusName(issue: Issue) = _uiState.value.statuses.firstOrNull { it.id == issue.statusId }?.name

    fun assigneeName(issue: Issue) = issue.assigneeId?.let { id ->
        _uiState.value.members.firstOrNull { it.id == id }?.displayName
    }

    fun updateIssue(title: String, statusId: String, assigneeId: String?, onSuccess: () -> Unit) {
        viewModelScope.launch {
            _uiState.update { it.copy(isSaving = true, saveError = null) }
            runCatching {
                api.updateIssue(
                    orgSlug,
                    projectKey,
                    issueNumber,
                    mapOf(
                        "title" to title,
                        "status_id" to statusId,
                        "assignee_id" to assigneeId,
                    ),
                )
            }.onSuccess {
                load()
                _uiState.update { it.copy(isSaving = false) }
                onSuccess()
            }.onFailure { e ->
                _uiState.update { it.copy(isSaving = false, saveError = e.message) }
            }
        }
    }

    fun clearSaveError() {
        _uiState.update { it.copy(saveError = null) }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun IssueDetailScreen(
    projectKey: String,
    onBack: () -> Unit = {},
    viewModel: IssueDetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsState()
    var commentText by remember { mutableStateOf("") }
    var showEditDialog by remember { mutableStateOf(false) }

    if (showEditDialog && uiState.issue != null) {
        val issue = uiState.issue!!
        EditIssueDialog(
            initialTitle = issue.title,
            initialStatusId = issue.statusId,
            initialAssigneeId = issue.assigneeId,
            statuses = uiState.statuses,
            members = uiState.members,
            isSubmitting = uiState.isSaving,
            error = uiState.saveError,
            onDismiss = {
                showEditDialog = false
                viewModel.clearSaveError()
            },
            onConfirm = { title, statusId, assigneeId ->
                viewModel.updateIssue(title, statusId, assigneeId) {
                    showEditDialog = false
                    viewModel.clearSaveError()
                }
            },
        )
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    uiState.issue?.let {
                        Text("$projectKey-${it.number}", style = MaterialTheme.typography.titleMedium)
                    }
                },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    if (uiState.issue != null) {
                        TextButton(onClick = { showEditDialog = true }) {
                            Text("Edit")
                        }
                    }
                },
            )
        },
        bottomBar = {
            Surface(shadowElevation = 8.dp) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp, vertical = 8.dp)
                        .navigationBarsPadding()
                        .imePadding(),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    OutlinedTextField(
                        value = commentText,
                        onValueChange = { commentText = it },
                        placeholder = { Text("Add a comment…") },
                        modifier = Modifier.weight(1f),
                        maxLines = 4,
                    )
                    Spacer(Modifier.width(8.dp))
                    IconButton(
                        onClick = {
                            viewModel.sendComment(commentText)
                            commentText = ""
                        },
                        enabled = commentText.isNotBlank() && !uiState.isSendingComment,
                    ) {
                        if (uiState.isSendingComment) {
                            CircularProgressIndicator(Modifier.size(20.dp), strokeWidth = 2.dp)
                        } else {
                            Icon(Icons.AutoMirrored.Filled.Send, contentDescription = "Send")
                        }
                    }
                }
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

            uiState.issue != null -> {
                val issue = uiState.issue!!
                LazyColumn(
                    modifier = Modifier.fillMaxSize().padding(padding),
                    contentPadding = PaddingValues(horizontal = 16.dp, vertical = 12.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp),
                ) {
                    item {
                        Text(issue.title, style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
                    }

                    item {
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            SuggestionChip(
                                onClick = { showEditDialog = true },
                                label = { Text(viewModel.statusName(issue) ?: "No status") },
                            )
                            SuggestionChip(
                                onClick = {},
                                label = { Text(issue.priority.replaceFirstChar { it.uppercase() }) },
                            )
                            issue.storyPoints?.let { sp ->
                                SuggestionChip(
                                    onClick = {},
                                    label = { Text("${sp.toInt()} pts") },
                                )
                            }
                        }
                    }

                    item {
                        Text(
                            "Assignee: ${viewModel.assigneeName(issue) ?: "Unassigned"}",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                        )
                    }

                    issue.description?.takeIf { it.isNotBlank() }?.let { desc ->
                        item {
                            Text("Description", style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                            Spacer(Modifier.height(4.dp))
                            Text(desc, style = MaterialTheme.typography.bodyMedium)
                        }
                    }

                    item {
                        HorizontalDivider()
                        Spacer(Modifier.height(4.dp))
                        Text(
                            "Comments (${uiState.comments.size})",
                            style = MaterialTheme.typography.titleSmall,
                            fontWeight = FontWeight.SemiBold,
                        )
                    }

                    items(uiState.comments, key = { it.id }) { comment ->
                        CommentCard(comment)
                    }

                    if (uiState.comments.isEmpty()) {
                        item {
                            Text(
                                "No comments yet.",
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun CommentCard(comment: Comment) {
    Card(
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant,
        ),
    ) {
        Column(modifier = Modifier.padding(12.dp)) {
            Row(
                horizontalArrangement = Arrangement.SpaceBetween,
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text(
                    comment.authorId.take(8),
                    style = MaterialTheme.typography.labelSmall,
                    fontWeight = FontWeight.SemiBold,
                    color = MaterialTheme.colorScheme.primary,
                )
                Text(
                    comment.createdAt.take(10),
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.outline,
                )
            }
            Spacer(Modifier.height(4.dp))
            Text(comment.body, style = MaterialTheme.typography.bodyMedium)
        }
    }
}
