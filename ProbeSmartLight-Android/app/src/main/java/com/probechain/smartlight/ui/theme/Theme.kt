// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.ui.theme

import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext

private val ProbeBlue = Color(0xFF2196F3)
private val ProbeGreen = Color(0xFF4CAF50)

private val DarkColorScheme = darkColorScheme(
    primary = ProbeBlue,
    secondary = ProbeGreen,
    tertiary = Color(0xFF03DAC5)
)

private val LightColorScheme = lightColorScheme(
    primary = ProbeBlue,
    secondary = ProbeGreen,
    tertiary = Color(0xFF018786)
)

@Composable
fun ProbeSmartLightTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    dynamicColor: Boolean = true,
    content: @Composable () -> Unit
) {
    val colorScheme = when {
        dynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            val context = LocalContext.current
            if (darkTheme) dynamicDarkColorScheme(context) else dynamicLightColorScheme(context)
        }
        darkTheme -> DarkColorScheme
        else -> LightColorScheme
    }

    MaterialTheme(
        colorScheme = colorScheme,
        content = content
    )
}
