// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.ui

import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.probechain.smartlight.ui.screens.*

@Composable
fun SmartLightNavHost() {
    val navController = rememberNavController()
    var selectedTab by remember { mutableIntStateOf(0) }

    val tabs = listOf(
        Triple("dashboard", "Dashboard", Icons.Default.Speed),
        Triple("wallet", "Wallet", Icons.Default.AccountBalanceWallet),
        Triple("agent", "Agent", Icons.Default.Memory),
        Triple("settings", "Settings", Icons.Default.Settings),
    )

    Scaffold(
        bottomBar = {
            NavigationBar {
                tabs.forEachIndexed { index, (route, label, icon) ->
                    NavigationBarItem(
                        icon = { Icon(icon, contentDescription = label) },
                        label = { Text(label) },
                        selected = selectedTab == index,
                        onClick = {
                            selectedTab = index
                            navController.navigate(route) {
                                popUpTo("dashboard") { saveState = true }
                                launchSingleTop = true
                                restoreState = true
                            }
                        }
                    )
                }
            }
        }
    ) { padding ->
        NavHost(
            navController = navController,
            startDestination = "dashboard",
            modifier = Modifier.padding(padding)
        ) {
            composable("dashboard") { DashboardScreen() }
            composable("wallet") { WalletScreen() }
            composable("agent") { AgentScreen() }
            composable("settings") { SettingsScreen() }
        }
    }
}
