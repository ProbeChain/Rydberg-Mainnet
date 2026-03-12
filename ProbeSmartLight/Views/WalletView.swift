// Copyright 2024 The ProbeChain Authors

import SwiftUI

struct WalletView: View {
    @EnvironmentObject var nodeService: NodeService
    @State private var showingKeyGen = false
    @State private var passphrase = ""
    @State private var walletAddress = ""
    @State private var balance = "0.00"
    @State private var stakeAmount = "10"
    @State private var isRegistered = false

    var body: some View {
        NavigationView {
            List {
                // Address Section
                Section(header: Text("SmartLight Address")) {
                    if walletAddress.isEmpty {
                        Button("Generate Dilithium Key") {
                            showingKeyGen = true
                        }
                    } else {
                        VStack(alignment: .leading, spacing: 4) {
                            Text("Address")
                                .font(.caption)
                                .foregroundColor(.secondary)
                            Text(walletAddress)
                                .font(.system(.caption, design: .monospaced))
                                .lineLimit(1)
                                .truncationMode(.middle)
                        }
                        HStack {
                            Text("Key Type")
                                .foregroundColor(.secondary)
                            Spacer()
                            Text("Dilithium (ML-DSA-44)")
                                .font(.caption)
                                .foregroundColor(.blue)
                        }
                    }
                }

                // Balance Section
                Section(header: Text("Balance")) {
                    HStack {
                        Text("PROBE")
                        Spacer()
                        Text("\(balance) PROBE")
                            .fontWeight(.semibold)
                    }
                }

                // Registration Section
                Section(header: Text("SmartLight Registration")) {
                    HStack {
                        Text("Status")
                        Spacer()
                        Text(isRegistered ? "Registered" : "Not Registered")
                            .foregroundColor(isRegistered ? .green : .orange)
                    }

                    HStack {
                        Text("Required Stake")
                        Spacer()
                        Text("10 PROBE")
                            .foregroundColor(.secondary)
                    }

                    if !isRegistered && !walletAddress.isEmpty {
                        Button("Register as SmartLight Node") {
                            register()
                        }
                        .foregroundColor(.blue)
                    }
                }

                // Security Section
                Section(header: Text("Key Security")) {
                    HStack {
                        Image(systemName: "lock.shield")
                        Text("Secure Enclave")
                        Spacer()
                        Text(SecureKeyWrapper.isSecureEnclaveAvailable ? "Available" : "Unavailable")
                            .foregroundColor(SecureKeyWrapper.isSecureEnclaveAvailable ? .green : .red)
                    }

                    HStack {
                        Image(systemName: "faceid")
                        Text("Biometric Lock")
                        Spacer()
                        Text("Enabled")
                            .foregroundColor(.green)
                    }

                    HStack {
                        Image(systemName: "key")
                        Text("Private Key Size")
                        Spacer()
                        Text("2528 bytes")
                            .foregroundColor(.secondary)
                    }
                }
            }
            .navigationTitle("Wallet")
            .sheet(isPresented: $showingKeyGen) {
                KeyGenSheet(passphrase: $passphrase, address: $walletAddress, isPresented: $showingKeyGen)
            }
        }
    }

    private func register() {
        // Registration will stake 10 PROBE and register the node
        isRegistered = true
    }
}

struct KeyGenSheet: View {
    @Binding var passphrase: String
    @Binding var address: String
    @Binding var isPresented: Bool
    @State private var confirmPassphrase = ""
    @State private var isGenerating = false
    @State private var errorMessage = ""

    var body: some View {
        NavigationView {
            Form {
                Section(header: Text("Create Dilithium Key"), footer: Text("Your quantum-resistant key will be encrypted with AES-256-GCM and protected by Face ID.")) {
                    SecureField("Passphrase", text: $passphrase)
                    SecureField("Confirm Passphrase", text: $confirmPassphrase)
                }

                if !errorMessage.isEmpty {
                    Section {
                        Text(errorMessage)
                            .foregroundColor(.red)
                    }
                }

                Section {
                    Button(action: generate) {
                        HStack {
                            if isGenerating {
                                ProgressView()
                            }
                            Text("Generate Key Pair")
                        }
                    }
                    .disabled(passphrase.isEmpty || passphrase != confirmPassphrase || isGenerating)
                }
            }
            .navigationTitle("New Key")
            .navigationBarItems(trailing: Button("Cancel") { isPresented = false })
        }
    }

    private func generate() {
        guard passphrase == confirmPassphrase else {
            errorMessage = "Passphrases don't match"
            return
        }
        isGenerating = true
        // Key generation happens via GoNodeBridge
        DispatchQueue.global().async {
            // Simulated — actual call goes through GoNodeBridge
            DispatchQueue.main.asyncAfter(deadline: .now() + 1) {
                address = "0x" + String(repeating: "a", count: 40) // placeholder
                isGenerating = false
                isPresented = false
            }
        }
    }
}
