package com.plexus.ui.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.plexus.data.api.PlexusApi
import com.plexus.data.models.User
import com.plexus.data.secure.SecureTokenStore
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

data class AuthUiState(
    val isLoading: Boolean = false,
    val isAuthenticated: Boolean = false,
    val currentUser: User? = null,
    val error: String? = null,
)

@HiltViewModel
class AuthViewModel @Inject constructor(
    private val api: PlexusApi,
    private val tokenStore: SecureTokenStore,
) : ViewModel() {

    private val _uiState = MutableStateFlow(AuthUiState())
    val uiState: StateFlow<AuthUiState> = _uiState.asStateFlow()

    init {
        if (tokenStore.getAccessToken() != null) {
            _uiState.update { it.copy(isAuthenticated = true) }
            viewModelScope.launch { fetchMe() }
        }
    }

    fun login(email: String, password: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                val pair = api.login(mapOf("email" to resolveLoginEmail(email), "password" to password))
                tokenStore.saveTokens(pair.accessToken, pair.refreshToken)
                _uiState.update { it.copy(isAuthenticated = true) }
                fetchMe()
            }.onFailure { e ->
                _uiState.update { it.copy(error = networkAwareMessage(e, "Login failed")) }
            }
            _uiState.update { it.copy(isLoading = false) }
        }
    }

    fun register(email: String, password: String, displayName: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                val pair = api.register(
                    mapOf("email" to email, "password" to password, "display_name" to displayName)
                )
                tokenStore.saveTokens(pair.accessToken, pair.refreshToken)
                _uiState.update { it.copy(isAuthenticated = true) }
                fetchMe()
            }.onFailure { e ->
                _uiState.update { it.copy(error = networkAwareMessage(e, "Registration failed")) }
            }
            _uiState.update { it.copy(isLoading = false) }
        }
    }

    fun logout() {
        tokenStore.clear()
        _uiState.update { AuthUiState() }
    }

    private suspend fun fetchMe() {
        runCatching { api.getMe() }.onSuccess { user ->
            _uiState.update { it.copy(currentUser = user) }
        }
    }

    companion object {
        /** Match web/iOS: shorthand `admin` → seed user email. */
        fun resolveLoginEmail(input: String): String {
            val trimmed = input.trim()
            return if (trimmed.equals("admin", ignoreCase = true)) "admin@plexus.local" else trimmed
        }

        private fun networkAwareMessage(e: Throwable, fallback: String): String =
            when (e) {
                is java.net.UnknownHostException,
                is java.net.ConnectException,
                is java.net.SocketTimeoutException,
                is java.io.IOException -> "Unable to connect to the server. Please try again."
                else -> e.message ?: fallback
            }
    }
}
