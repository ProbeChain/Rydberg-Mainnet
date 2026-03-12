// Copyright 2024 The ProbeChain Authors

import SwiftUI

struct ContentView: View {
    @EnvironmentObject var nodeService: NodeService

    var body: some View {
        TabView {
            DashboardView()
                .tabItem {
                    Image(systemName: "gauge.medium")
                    Text("Dashboard")
                }

            WalletView()
                .tabItem {
                    Image(systemName: "wallet.pass")
                    Text("Wallet")
                }

            AgentView()
                .tabItem {
                    Image(systemName: "cpu")
                    Text("Agent")
                }

            SettingsView()
                .tabItem {
                    Image(systemName: "gearshape")
                    Text("Settings")
                }
        }
        .accentColor(.blue)
    }
}
