package com.plexus.app

import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext

private val PlexusBlue = Color(0xFF0052CC)
private val PlexusBlueLight = Color(0xFF4C9AFF)

private val LightColors = lightColorScheme(
    primary = PlexusBlue,
    onPrimary = Color.White,
    primaryContainer = Color(0xFFDEEBFF),
    onPrimaryContainer = Color(0xFF0747A6),
    secondary = Color(0xFF5E6C84),
    background = Color(0xFFF4F5F7),
    surface = Color.White,
    surfaceVariant = Color(0xFFEBECF0),
    onSurface = Color(0xFF172B4D),
    onSurfaceVariant = Color(0xFF5E6C84),
    outline = Color(0xFF8993A4),
)

private val DarkColors = darkColorScheme(
    primary = PlexusBlueLight,
    onPrimary = Color(0xFF172B4D),
    primaryContainer = Color(0xFF0747A6),
    onPrimaryContainer = Color(0xFFDEEBFF),
    secondary = Color(0xFF9FADBC),
    background = Color(0xFF1D2125),
    surface = Color(0xFF22272B),
    surfaceVariant = Color(0xFF2C333A),
    onSurface = Color(0xFFDEE4EA),
    onSurfaceVariant = Color(0xFF9FADBC),
    outline = Color(0xFF738496),
)

@Composable
fun PlexusTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    dynamicColor: Boolean = false,
    content: @Composable () -> Unit,
) {
    val colorScheme = when {
        dynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            val context = LocalContext.current
            if (darkTheme) dynamicDarkColorScheme(context) else dynamicLightColorScheme(context)
        }
        darkTheme -> DarkColors
        else -> LightColors
    }

    MaterialTheme(
        colorScheme = colorScheme,
        typography = Typography(),
        content = content,
    )
}
