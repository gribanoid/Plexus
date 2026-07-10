package com.plexus.app

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.*
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.plexus.ui.auth.AuthViewModel
import com.plexus.ui.auth.LoginScreen
import com.plexus.ui.auth.RegisterScreen
import com.plexus.ui.backlog.BacklogScreen
import com.plexus.ui.board.BoardScreen
import com.plexus.ui.issue.IssueDetailScreen
import com.plexus.ui.notifications.NotificationsScreen
import com.plexus.ui.orgs.OrgsScreen
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

@Composable
fun PlexusNavHost() {
    val navController = rememberNavController()
    val authViewModel: AuthViewModel = hiltViewModel()
    val authState by authViewModel.uiState.collectAsState()

    // Redirect based on auth state once determined (only on first composition)
    val startDestination = if (authState.isAuthenticated) "orgs" else "login"

    NavHost(
        navController = navController,
        startDestination = startDestination,
    ) {
        composable("login") {
            LoginScreen(
                onSuccess = {
                    navController.navigate("orgs") {
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
                    navController.navigate("orgs") {
                        popUpTo("register") { inclusive = true }
                        popUpTo("login") { inclusive = true }
                    }
                },
                onNavigateToLogin = { navController.popBackStack() },
                viewModel = authViewModel,
            )
        }

        composable("orgs") {
            OrgsScreen(
                onOrgSelected = { orgSlug ->
                    navController.navigate("projects/$orgSlug")
                },
                onNotifications = {
                    navController.navigate("notifications")
                },
                onLogout = {
                    authViewModel.logout()
                    navController.navigate("login") {
                        popUpTo(0) { inclusive = true }
                    }
                },
            )
        }

        composable("notifications") {
            NotificationsScreen(onBack = { navController.popBackStack() })
        }

        composable(
            route = "projects/{orgSlug}",
            arguments = listOf(navArgument("orgSlug") { type = NavType.StringType }),
        ) { backStack ->
            val orgSlug = backStack.arguments!!.getString("orgSlug")!!
            ProjectsScreen(
                orgSlug = orgSlug,
                onProjectSelected = { projectKey ->
                    navController.navigate("board/$orgSlug/$projectKey")
                },
                onBacklogSelected = { projectKey ->
                    navController.navigate("backlog/$orgSlug/$projectKey")
                },
                onBack = { navController.popBackStack() },
            )
        }

        composable(
            route = "board/{orgSlug}/{projectKey}",
            arguments = listOf(
                navArgument("orgSlug") { type = NavType.StringType },
                navArgument("projectKey") { type = NavType.StringType },
            ),
        ) { backStack ->
            val orgSlug = backStack.arguments!!.getString("orgSlug")!!
            val projectKey = backStack.arguments!!.getString("projectKey")!!
            BoardScreen(
                projectKey = projectKey,
                onBack = { navController.popBackStack() },
                onIssueClick = { issueNumber ->
                    navController.navigate("issue/$orgSlug/$projectKey/$issueNumber")
                },
            )
        }

        composable(
            route = "backlog/{orgSlug}/{projectKey}",
            arguments = listOf(
                navArgument("orgSlug") { type = NavType.StringType },
                navArgument("projectKey") { type = NavType.StringType },
            ),
        ) { backStack ->
            val orgSlug = backStack.arguments!!.getString("orgSlug")!!
            val projectKey = backStack.arguments!!.getString("projectKey")!!
            BacklogScreen(
                projectKey = projectKey,
                onBack = { navController.popBackStack() },
                onIssueClick = { issueNumber ->
                    navController.navigate("issue/$orgSlug/$projectKey/$issueNumber")
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
        ) { backStack ->
            val projectKey = backStack.arguments!!.getString("projectKey")!!
            IssueDetailScreen(projectKey = projectKey, onBack = { navController.popBackStack() })
        }
    }
}
