// Copyright 2024 The ProbeChain Authors

import SwiftUI

struct AgentView: View {
    @EnvironmentObject var nodeService: NodeService
    @State private var tasksCompleted: UInt64 = 0
    @State private var tasksSucceeded: UInt64 = 0
    @State private var activeTasks: Int = 0

    var body: some View {
        NavigationView {
            List {
                // Agent Status
                Section(header: Text("Agent Task Runner")) {
                    HStack {
                        Text("Status")
                        Spacer()
                        Text(nodeService.isRunning ? "Active" : "Inactive")
                            .foregroundColor(nodeService.isRunning ? .green : .gray)
                    }

                    HStack {
                        Text("Max Concurrent Tasks")
                        Spacer()
                        Text("2")
                            .foregroundColor(.secondary)
                    }

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

                // Task Statistics
                Section(header: Text("Statistics")) {
                    HStack {
                        Text("Active Tasks")
                        Spacer()
                        Text("\(activeTasks) / 2")
                            .foregroundColor(.blue)
                    }

                    HStack {
                        Text("Tasks Completed")
                        Spacer()
                        Text("\(tasksCompleted)")
                    }

                    HStack {
                        Text("Success Rate")
                        Spacer()
                        let rate = tasksCompleted > 0
                            ? Double(tasksSucceeded) / Double(tasksCompleted) * 100
                            : 100.0
                        Text(String(format: "%.1f%%", rate))
                            .foregroundColor(rate >= 90 ? .green : .orange)
                    }
                }

                // Power Mode Impact
                Section(header: Text("Power Mode Impact")) {
                    HStack {
                        Image(systemName: "bolt.fill")
                            .foregroundColor(.yellow)
                        Text("Full Mode")
                        Spacer()
                        Text("Tasks Enabled")
                            .foregroundColor(.green)
                            .font(.caption)
                    }

                    HStack {
                        Image(systemName: "leaf.fill")
                            .foregroundColor(.green)
                        Text("Eco Mode")
                        Spacer()
                        Text("Tasks Disabled")
                            .foregroundColor(.orange)
                            .font(.caption)
                    }

                    HStack {
                        Image(systemName: "moon.fill")
                            .foregroundColor(.indigo)
                        Text("Sleep Mode")
                        Spacer()
                        Text("Tasks Disabled")
                            .foregroundColor(.red)
                            .font(.caption)
                    }
                }

                // Info
                Section(footer: Text("Agent tasks are lightweight computations requested by the ProbeChain network. Tasks execute in a sandboxed environment with strict resource limits. Only available in Full power mode (while charging).")) {
                    EmptyView()
                }
            }
            .navigationTitle("Agent")
        }
    }
}
