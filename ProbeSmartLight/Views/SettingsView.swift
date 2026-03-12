// Copyright 2024 The ProbeChain Authors

import SwiftUI

struct SettingsView: View {
    @EnvironmentObject var nodeService: NodeService
    @State private var gnssEnabled = true
    @State private var heartbeatInterval = 100
    @State private var maxAgentTasks = 2
    @State private var selectedPowerMode = 0
    @State private var showingResetAlert = false

    let powerModes = ["Full (Charging)", "Eco (Battery > 30%)", "Sleep (Battery < 15%)"]

    var body: some View {
        NavigationView {
            Form {
                // Node Settings
                Section(header: Text("Node")) {
                    HStack {
                        Text("Network ID")
                        Spacer()
                        Text("8004 (PoB ERC-8004)")
                            .foregroundColor(.secondary)
                    }

                    HStack {
                        Text("Database Cache")
                        Spacer()
                        Text("64 MB")
                            .foregroundColor(.secondary)
                    }

                    HStack {
                        Text("Max Peers")
                        Spacer()
                        Text("50")
                            .foregroundColor(.secondary)
                    }
                }

                // Power Management
                Section(header: Text("Power Management")) {
                    Picker("Power Mode", selection: $selectedPowerMode) {
                        ForEach(0..<powerModes.count, id: \.self) { index in
                            Text(powerModes[index]).tag(index)
                        }
                    }
                    .onChange(of: selectedPowerMode) { newValue in
                        nodeService.setPowerMode(newValue)
                    }

                    VStack(alignment: .leading, spacing: 4) {
                        Text("Estimated Battery Usage")
                            .font(.caption)
                            .foregroundColor(.secondary)
                        switch selectedPowerMode {
                        case 0: Text("~5% / hour").foregroundColor(.orange)
                        case 1: Text("~2% / hour").foregroundColor(.green)
                        default: Text("~0.5% / hour").foregroundColor(.green)
                        }
                    }
                }

                // GNSS Settings
                Section(header: Text("GNSS Time Source")) {
                    Toggle("Enable GPS Time Sampling", isOn: $gnssEnabled)

                    if gnssEnabled {
                        HStack {
                            Text("Sample Interval")
                            Spacer()
                            Text("30 seconds")
                                .foregroundColor(.secondary)
                        }

                        HStack {
                            Text("Accuracy")
                            Spacer()
                            Text("~100 ns")
                                .foregroundColor(.secondary)
                        }
                    }
                }

                // Heartbeat Settings
                Section(header: Text("Heartbeat")) {
                    HStack {
                        Text("Interval")
                        Spacer()
                        Text("Every \(heartbeatInterval) blocks")
                            .foregroundColor(.secondary)
                    }
                }

                // Agent Settings
                Section(header: Text("Agent Tasks")) {
                    Stepper("Max Tasks: \(maxAgentTasks)", value: $maxAgentTasks, in: 1...4)

                    HStack {
                        Text("Max Memory per Task")
                        Spacer()
                        Text("25 MB")
                            .foregroundColor(.secondary)
                    }

                    HStack {
                        Text("Task Timeout")
                        Spacer()
                        Text("5 seconds")
                            .foregroundColor(.secondary)
                    }
                }

                // Security
                Section(header: Text("Security")) {
                    HStack {
                        Text("Quantum-Safe Keys")
                        Spacer()
                        Text("CRYSTALS-Dilithium")
                            .foregroundColor(.blue)
                            .font(.caption)
                    }

                    HStack {
                        Text("Key Protection")
                        Spacer()
                        Text("Secure Enclave + AES-256-GCM")
                            .foregroundColor(.blue)
                            .font(.caption)
                    }

                    HStack {
                        Text("Anti-Sybil Stake")
                        Spacer()
                        Text("10 PROBE")
                            .foregroundColor(.secondary)
                    }
                }

                // About
                Section(header: Text("About")) {
                    HStack {
                        Text("Version")
                        Spacer()
                        Text("2.0.0-smartlight")
                            .foregroundColor(.secondary)
                    }

                    HStack {
                        Text("ProbeChain")
                        Spacer()
                        Text("V2.0 PoB Rydberg + StellarSpeed")
                            .foregroundColor(.secondary)
                    }

                    HStack {
                        Text("Genesis Node")
                        Spacer()
                        Text("192.168.110.142:30303")
                            .foregroundColor(.secondary)
                            .font(.caption)
                    }
                }

                // Reset
                Section {
                    Button("Reset Node Data") {
                        showingResetAlert = true
                    }
                    .foregroundColor(.red)
                }
            }
            .navigationTitle("Settings")
            .alert("Reset Node Data", isPresented: $showingResetAlert) {
                Button("Cancel", role: .cancel) {}
                Button("Reset", role: .destructive) {
                    // Reset node data
                }
            } message: {
                Text("This will delete all synced data and require re-sync. Your keys will be preserved.")
            }
        }
    }
}
