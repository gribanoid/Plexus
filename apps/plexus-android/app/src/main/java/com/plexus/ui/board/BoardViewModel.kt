package com.plexus.ui.board

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.plexus.data.api.PlexusApi
import com.plexus.data.models.Issue
import com.plexus.data.models.IssueType
import com.plexus.data.models.OrgMember
import com.plexus.data.models.Status
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.async
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

data class BoardUiState(
    val isLoading: Boolean = false,
    val statuses: List<Status> = emptyList(),
    val issues: List<Issue> = emptyList(),
    val issueTypes: List<IssueType> = emptyList(),
    val members: List<OrgMember> = emptyList(),
    val isLoadingIssueTypes: Boolean = false,
    val isCreating: Boolean = false,
    val createError: String? = null,
    val error: String? = null,
) {
    fun issuesFor(statusId: String) = issues
        .filter { it.statusId == statusId }
        .sortedBy { it.position }
}

@HiltViewModel
class BoardViewModel @Inject constructor(
    private val api: PlexusApi,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {

    private val orgSlug: String = checkNotNull(savedStateHandle["orgSlug"])
    private val projectKey: String = checkNotNull(savedStateHandle["projectKey"])

    private val _uiState = MutableStateFlow(BoardUiState(isLoading = true))
    val uiState: StateFlow<BoardUiState> = _uiState.asStateFlow()

    init {
        load()
    }

    fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                val statusDef = async { api.listStatuses(orgSlug, projectKey) }
                val issueDef = async { api.listIssues(orgSlug, projectKey) }
                val typesDef = async { api.listIssueTypes(orgSlug, projectKey) }
                val membersDef = async { api.listOrgMembers(orgSlug) }
                val statuses = statusDef.await().items.sortedBy { it.position }
                val issues = issueDef.await().items
                val types = typesDef.await().items
                val members = membersDef.await().items
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        statuses = statuses,
                        issues = issues,
                        issueTypes = types,
                        members = members,
                    )
                }
            }.onFailure { e ->
                _uiState.update { it.copy(isLoading = false, error = e.message) }
            }
        }
    }

    fun moveIssue(issueNumber: Int, newStatusId: String) {
        viewModelScope.launch {
            runCatching {
                api.moveIssue(orgSlug, projectKey, issueNumber, mapOf("status_id" to newStatusId))
                _uiState.update { state ->
                    state.copy(
                        issues = state.issues.map { issue ->
                            if (issue.number == issueNumber) issue.copy(statusId = newStatusId)
                            else issue
                        }
                    )
                }
            }
        }
    }

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

    fun createIssue(
        title: String,
        typeId: String,
        priority: String,
        statusId: String? = null,
        onSuccess: () -> Unit,
    ) {
        viewModelScope.launch {
            _uiState.update { it.copy(isCreating = true, createError = null) }
            runCatching {
                val body = mutableMapOf<String, Any>(
                    "title" to title,
                    "type_id" to typeId,
                    "priority" to priority,
                )
                if (statusId != null) body["status_id"] = statusId
                api.createIssue(orgSlug, projectKey, body)
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
