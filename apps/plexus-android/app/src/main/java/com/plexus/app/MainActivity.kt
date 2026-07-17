package com.plexus.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AccountCircle
import androidx.compose.material.icons.filled.Assignment
import androidx.compose.material.icons.filled.Folder
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavGraph.Companion.findStartDestination
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.plexus.ui.account.AccountScreen
import com.plexus.ui.auth.AuthViewModel
import com.plexus.ui.auth.LoginScreen
import com.plexus.ui.auth.RegisterScreen
import com.plexus.ui.issue.IssueDetailScreen
import com.plexus.ui.issues.IssuesListScreen
import com.plexus.ui.notifications.NotificationsScreen
import com.plexus.ui.orgs.OrgsScreen
import com.plexus.ui.project.ProjectHubScreen
import com.plexus.ui.projects.ProjectsScreen
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()
        setContent {
            PlexusTheme {
                Surface(color = MaterialTheme.colorScheme.background) {
                    PlexusNavHost()
                }
            }
        }
    }
}

private data class BottomDest(
    val route: String,
    val label: String,
    val icon: androidx.compose.ui.graphics.vector.ImageVector,
)

private val bottomDestinations = listOf(
    BottomDest("projects_tab", "Projects", Icons.Default.Folder),
    BottomDest("issues_tab", "Issues", Icons.Default.Assignment),
    BottomDest("notifications_tab", "Notifications", Icons.Default.Notifications),
    BottomDest("account_tab", "Account", Icons.Default.AccountCircle),
)

@Composable
fun PlexusNavHost() {
    val navController = rememberNavController()
    val authViewModel: AuthViewModel = hiltViewModel()
    val authState by authViewModel.uiState.collectAsState()

    val startDestination = if (authState.isAuthenticated) "main" else "login"

    NavHost(
        navController = navController,
        startDestination = startDestination,
    ) {
        composable("login") {
            LoginScreen(
                onSuccess = {
                    navController.navigate("main") {
                        popUpTo("login") { inclusive = true }
                    }
                },
                onNavigateToRegister = { navController.navigate("register") },
                viewModel = authViewModel,
            )
        }

        composable("register") {
            RegisterScreen(
                onSuccess = {
                    navController.navigate("main") {
                        popUpTo("register") { inclusive = true }
                        popUpTo("login") { inclusive = true }
                    }
                },
                onNavigateToLogin = { navController.popBackStack() },
                viewModel = authViewModel,
            )
        }

        composable("main") {
            MainShell(
                authViewModel = authViewModel,
                onLogout = {
                    navController.navigate("login") {
                        popUpTo(0) { inclusive = true }
                    }
                },
            )
        }
    }
}

@Composable
fun MainShell(
    authViewModel: AuthViewModel,
    onLogout: () -> Unit,
) {
    val tabNavController = rememberNavController()
    val navBackStackEntry by tabNavController.currentBackStackEntryAsState()
    val currentRoute = navBackStackEntry?.destination?.route
    val showBottomBar = currentRoute?.startsWith("issue/") != true

    Scaffold(
        bottomBar = {
            if (showBottomBar) {
                NavigationBar {
                    bottomDestinations.forEach { dest ->
                        val selected = when {
                            dest.route == "projects_tab" ->
                                currentRoute == "projects_tab" ||
                                    currentRoute == "orgs" ||
                                    currentRoute?.startsWith("projects/") == true ||
                                    currentRoute?.startsWith("hub/") == true
                            else -> currentRoute == dest.route
                        }
                        NavigationBarItem(
                            selected = selected,
                            onClick = {
                                tabNavController.navigate(dest.route) {
                                    popUpTo(tabNavController.graph.findStartDestination().id) {
                                        saveState = true
                                    }
                                    launchSingleTop = true
                                    restoreState = true
                                }
                            },
                            icon = { Icon(dest.icon, contentDescription = dest.label) },
                            label = { Text(dest.label) },
                        )
                    }
                }
            }
        },
    ) { padding ->
        NavHost(
            navController = tabNavController,
            startDestination = "projects_tab",
            modifier = Modifier.padding(padding),
        ) {
            composable("projects_tab") {
                OrgsScreen(
                    onOrgSelected = { orgSlug ->
                        tabNavController.navigate("projects/$orgSlug")
                    },
                )
            }

            composable("issues_tab") {
                IssuesListScreen(
                    onIssueClick = { orgSlug, projectKey, issueNumber ->
                        tabNavController.navigate("issue/$orgSlug/$projectKey/$issueNumber")
                    },
                )
            }

            composable("notifications_tab") {
                NotificationsScreen(showBack = false)
            }

            composable("account_tab") {
                AccountScreen(
                    onLogout = {
                        authViewModel.logout()
                        onLogout()
                    },
                    authViewModel = authViewModel,
                )
            }

            composable(
                route = "projects/{orgSlug}",
                arguments = listOf(navArgument("orgSlug") { type = NavType.StringType }),
            ) { backStack ->
                val orgSlug = backStack.arguments!!.getString("orgSlug")!!
                ProjectsScreen(
                    orgSlug = orgSlug,
                    onProjectSelected = { projectKey ->
                        tabNavController.navigate("hub/$orgSlug/$projectKey")
                    },
                    onBack = { tabNavController.popBackStack() },
                )
            }

            composable(
                route = "hub/{orgSlug}/{projectKey}",
                arguments = listOf(
                    navArgument("orgSlug") { type = NavType.StringType },
                    navArgument("projectKey") { type = NavType.StringType },
                ),
            ) { backStack ->
                val orgSlug = backStack.arguments!!.getString("orgSlug")!!
                val projectKey = backStack.arguments!!.getString("projectKey")!!
                ProjectHubScreen(
                    orgSlug = orgSlug,
                    projectKey = projectKey,
                    onBack = { tabNavController.popBackStack() },
                    onIssueClick = { issueNumber ->
                        tabNavController.navigate("issue/$orgSlug/$projectKey/$issueNumber")
                    },
                )
            }

            composable(
                route = "issue/{orgSlug}/{projectKey}/{issueNumber}",
                arguments = listOf(
                    navArgument("orgSlug") { type = NavType.StringType },
                    navArgument("projectKey") { type = NavType.StringType },
                    navArgument("issueNumber") { type = NavType.StringType },
                ),
            ) {
                val projectKey = it.arguments!!.getString("projectKey")!!
                IssueDetailScreen(projectKey = projectKey, onBack = { tabNavController.popBackStack() })
            }
        }
    }
}
