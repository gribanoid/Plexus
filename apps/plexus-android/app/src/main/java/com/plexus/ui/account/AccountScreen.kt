package com.plexus.ui.account

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.plexus.ui.auth.AuthViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AccountScreen(
    onLogout: () -> Unit,
    authViewModel: AuthViewModel = hiltViewModel(),
) {
    val authState by authViewModel.uiState.collectAsState()
    val user = authState.currentUser

    Scaffold(
        topBar = {
            TopAppBar(title = { Text("Account") })
        },
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(24.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            if (user != null) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(16.dp),
                ) {
                    Surface(
                        shape = MaterialTheme.shapes.extraLarge,
                        color = MaterialTheme.colorScheme.primary,
                        modifier = Modifier.size(56.dp),
                    ) {
                        Box(contentAlignment = Alignment.Center) {
                            Text(
                                user.displayName.take(1).uppercase(),
                                style = MaterialTheme.typography.titleLarge,
                                color = MaterialTheme.colorScheme.onPrimary,
                                fontWeight = FontWeight.Bold,
                            )
                        }
                    }
                    Column {
                        Text(user.displayName, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
                        Text(user.email, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
            } else {
                Text("Loading profile…", color = MaterialTheme.colorScheme.onSurfaceVariant)
            }

            Spacer(Modifier.weight(1f))

            Button(
                onClick = {
                    authViewModel.logout()
                    onLogout()
                },
                colors = ButtonDefaults.buttonColors(
                    containerColor = MaterialTheme.colorScheme.error,
                ),
                modifier = Modifier.fillMaxWidth(),
            ) {
                Text("Sign out")
            }
        }
    }
}
