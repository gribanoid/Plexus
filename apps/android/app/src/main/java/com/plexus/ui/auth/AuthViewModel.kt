package com.plexus.ui.auth

import android.content.Context
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.plexus.data.api.PlexusApi
import com.plexus.data.models.User
import com.plexus.di.tokenDataStore
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import javax.inject.Inject

val KEY_ACCESS_TOKEN = stringPreferencesKey("access_token")
val KEY_REFRESH_TOKEN = stringPreferencesKey("refresh_token")

data class AuthUiState(
    val isLoading: Boolean = false,
    val isAuthenticated: Boolean = false,
    val currentUser: User? = null,
    val error: String? = null,
)

@HiltViewModel
class AuthViewModel @Inject constructor(
    private val api: PlexusApi,
    @ApplicationContext private val context: Context,
) : ViewModel() {

    private val _uiState = MutableStateFlow(AuthUiState())
    val uiState: StateFlow<AuthUiState> = _uiState.asStateFlow()

    init {
        // Check if we already have a stored token
        viewModelScope.launch {
            context.tokenDataStore.data.first().let { prefs ->
                val token = prefs[KEY_ACCESS_TOKEN]
                if (token != null) {
                    _uiState.update { it.copy(isAuthenticated = true) }
                    fetchMe()
                }
            }
        }
    }

    fun login(email: String, password: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }
            runCatching {
                val pair = api.login(mapOf("email" to email, "password" to password))
                saveTokens(pair.accessToken, pair.refreshToken)
                _uiState.update { it.copy(isAuthenticated = true) }
                fetchMe()
            }.onFailure { e ->
                _uiState.update { it.copy(error = e.message ?: "Login failed") }
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
                saveTokens(pair.accessToken, pair.refreshToken)
                _uiState.update { it.copy(isAuthenticated = true) }
                fetchMe()
            }.onFailure { e ->
                _uiState.update { it.copy(error = e.message ?: "Registration failed") }
            }
            _uiState.update { it.copy(isLoading = false) }
        }
    }

    fun logout() {
        viewModelScope.launch {
            context.tokenDataStore.edit { it.clear() }
            _uiState.update { AuthUiState() }
        }
    }

    private suspend fun fetchMe() {
        runCatching { api.getMe() }.onSuccess { user ->
            _uiState.update { it.copy(currentUser = user) }
        }
    }

    private suspend fun saveTokens(accessToken: String, refreshToken: String) {
        context.tokenDataStore.edit { prefs ->
            prefs[KEY_ACCESS_TOKEN] = accessToken
            prefs[KEY_REFRESH_TOKEN] = refreshToken
        }
    }
}
