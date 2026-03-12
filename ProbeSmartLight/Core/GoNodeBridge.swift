// Copyright 2024 The ProbeChain Authors
// GoNodeBridge — Swift wrapper for the gomobile-generated Gprobe framework.
// Bridges SmartLightNode Go methods to Swift async/await.

import Foundation
import Gprobe  // gomobile-generated .xcframework

/// GoNodeBridge wraps the Go SmartLight node for Swift consumption.
final class GoNodeBridge: ObservableObject {
    private var node: GprobeSmartLightNode?
    private var keystore: GprobeDilithiumKeyStore?

    @Published var isRunning = false
    @Published var syncedBlock: UInt64 = 0
    @Published var peerCount: Int = 0

    // MARK: - Lifecycle

    // MARK: - Network Constants

    /// PoB chain network ID (Rydberg Upgrade)
    static let pobNetworkID: Int64 = 8004

    /// Genesis node enode URL
    static let genesisEnode = "enode://59a7202485ca9e6067cb2d9f1071ac3c7258c9450648ee6bcfab07ba14533794a9282a7b287fdec904e15e3bb45e51c439eecd4d7904c6ebffb5d676f9e27107@192.168.110.142:30303"

    /// PoB Genesis JSON
    static let pobGenesisJSON = """
    {"config":{"chainId":8004,"homesteadBlock":0,"eip150Block":0,"eip150Hash":"0x0000000000000000000000000000000000000000000000000000000000000000","eip155Block":0,"eip158Block":0,"byzantiumBlock":0,"constantinopleBlock":0,"petersburgBlock":0,"istanbulBlock":0,"muirGlacierBlock":0,"berlinBlock":0,"londonBlock":0,"shenzhenBlock":0,"stellarSpeedBlock":0,"pob":{"period":0,"tickIntervalMs":400,"epoch":30000,"initialScore":5000,"slashFraction":1000,"demotionThreshold":1000,"list":[{"enode":"enode://59a7202485ca9e6067cb2d9f1071ac3c7258c9450648ee6bcfab07ba14533794a9282a7b287fdec904e15e3bb45e51c439eecd4d7904c6ebffb5d676f9e27107@127.0.0.1:30303","owner":"0x63b3567834350660acaf54a7773d4bcce731b982"}]}},"nonce":"0x0","timestamp":"0x65f18600","extraData":"0x0000000000000000000000000000000000000000000000000000000000000000","gasLimit":"0x1c9c380","difficulty":"0x1","mixHash":"0x0000000000000000000000000000000000000000000000000000000000000000","coinbase":"0x0000000000000000000000000000000000000000","alloc":{"0x63b3567834350660acaf54a7773d4bcce731b982":{"balance":"1000000000000000000000000000"}},"number":"0x0","gasUsed":"0x0","parentHash":"0x0000000000000000000000000000000000000000000000000000000000000000","baseFeePerGas":null}
    """

    /// Initializes and starts the SmartLight node.
    func start(dataDir: String, networkID: Int64 = GoNodeBridge.pobNetworkID) throws {
        let config = GprobeNewSmartLightNodeConfig()
        config?.probeumNetworkID = networkID
        config?.probeumDatabaseCache = 64
        config?.heartbeatInterval = 100
        config?.gnssEnabled = true
        config?.maxAgentTasks = 2
        config?.powerMode = 0 // Full

        // Set PoB genesis
        config?.probeumGenesis = GoNodeBridge.pobGenesisJSON

        // Add genesis bootnode
        var enodeError: NSError?
        if let genesisNode = GprobeNewEnode(GoNodeBridge.genesisEnode, &enodeError) {
            let bootnodes = GprobeNewEnodesEmpty()
            bootnodes?.append(genesisNode)
            config?.bootstrapNodes = bootnodes
        }

        var error: NSError?
        node = GprobeNewSmartLightNode(dataDir, config, &error)
        if let error = error {
            throw error
        }

        try node?.start()
        isRunning = true

        // Initialize Dilithium keystore
        let keyDir = (dataDir as NSString).appendingPathComponent("dilithium-keys")
        keystore = GprobeNewDilithiumKeyStore(keyDir)
    }

    /// Stops the SmartLight node.
    func stop() throws {
        try node?.close()
        isRunning = false
        node = nil
    }

    // MARK: - Node Info

    /// Returns the current peer count.
    func getPeerCount() -> Int {
        return node?.getPeersInfo()?.size() ?? 0
    }

    /// Returns the current synced block number.
    func getSyncedBlockNumber() -> UInt64 {
        return UInt64(node?.getSyncedBlockNumber() ?? 0)
    }

    /// Returns the node's enode URL.
    func getEnode() -> String {
        return node?.getInfo()?.getEnode() ?? ""
    }

    // MARK: - Power Mode

    /// Sets the power mode: 0=Full, 1=Eco, 2=Sleep.
    func setPowerMode(_ mode: Int) {
        node?.setPowerMode(mode)
    }

    /// Gets the current power mode.
    func getPowerMode() -> Int {
        return node?.getPowerMode() ?? 0
    }

    // MARK: - Behavior Score

    /// Returns the JSON-encoded behavior score.
    func getBehaviorScore() -> String {
        var error: NSError?
        let json = node?.getBehaviorScore(&error) ?? "{}"
        return json
    }

    /// Returns the JSON-encoded reward stats.
    func getRewardStats() -> String {
        var error: NSError?
        let json = node?.getRewardStats(&error) ?? "{}"
        return json
    }

    // MARK: - Dilithium Key Management

    /// Generates a new Dilithium key pair.
    func generateDilithiumKey(passphrase: String) throws -> String {
        let info = try keystore?.generateKey(passphrase)
        return info?.address ?? ""
    }

    /// Signs a message with the Dilithium key.
    func signWithDilithium(address: String, passphrase: String, message: Data) throws -> Data {
        guard let privKey = try keystore?.loadKey(address, passphrase: passphrase) else {
            throw NSError(domain: "GoNodeBridge", code: -1, userInfo: [NSLocalizedDescriptionKey: "Key not found"])
        }
        let sig = try keystore?.sign(privKey, message: message)
        return sig ?? Data()
    }

    /// Lists all stored Dilithium keys.
    func listDilithiumKeys() -> String {
        var error: NSError?
        return keystore?.listKeys(&error) ?? "[]"
    }
}
