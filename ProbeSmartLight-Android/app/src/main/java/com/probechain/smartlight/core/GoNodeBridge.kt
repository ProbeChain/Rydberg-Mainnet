// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.core

import android.content.Context
import gprobe.Gprobe
import gprobe.SmartLightNode
import gprobe.SmartLightNodeConfig
import gprobe.DilithiumKeyStore
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import org.json.JSONObject
import java.io.File

/**
 * GoNodeBridge wraps the gomobile-generated Gprobe library for Kotlin consumption.
 * Bridges SmartLightNode Go methods to Kotlin coroutines.
 */
class GoNodeBridge(private val context: Context) {

    private var node: SmartLightNode? = null
    private var keystore: DilithiumKeyStore? = null

    private val _isRunning = MutableStateFlow(false)
    val isRunning: StateFlow<Boolean> = _isRunning.asStateFlow()

    private val _syncedBlock = MutableStateFlow(0L)
    val syncedBlock: StateFlow<Long> = _syncedBlock.asStateFlow()

    private val _peerCount = MutableStateFlow(0)
    val peerCount: StateFlow<Int> = _peerCount.asStateFlow()

    // MARK: - Lifecycle

    /**
     * Initializes and starts the SmartLight node.
     */
    fun start(networkID: Long = 1205) {
        val dataDir = getDataDir()
        val config = Gprobe.newSmartLightNodeConfig().apply {
            probeumNetworkID = networkID
            probeumDatabaseCache = 64
            heartbeatInterval = 100
            gnssEnabled = true
            maxAgentTasks = 2
            powerMode = 0 // Full
        }

        node = Gprobe.newSmartLightNode(dataDir, config)
        node?.start()
        _isRunning.value = true

        // Initialize Dilithium keystore
        val keyDir = File(dataDir, "dilithium-keys").absolutePath
        keystore = Gprobe.newDilithiumKeyStore(keyDir)
    }

    /**
     * Stops the SmartLight node.
     */
    fun stop() {
        node?.close()
        _isRunning.value = false
        node = null
    }

    // MARK: - Node Info

    fun getPeerCount(): Int = node?.peersInfo?.size()?.toInt() ?: 0

    fun getEnode(): String = node?.nodeInfo?.enode ?: ""

    // MARK: - Power Mode

    fun setPowerMode(mode: Int) {
        node?.setPowerMode(mode.toLong())
    }

    fun getPowerMode(): Int = node?.powerMode?.toInt() ?: 0

    // MARK: - Behavior Score

    fun getBehaviorScore(): BehaviorScoreData {
        val json = try { node?.behaviorScore ?: "{}" } catch (_: Exception) { "{}" }
        return parseBehaviorScore(json)
    }

    fun getRewardStats(): RewardStatsData {
        val json = try { node?.rewardStats ?: "{}" } catch (_: Exception) { "{}" }
        return parseRewardStats(json)
    }

    // MARK: - Dilithium Key Management

    fun generateDilithiumKey(passphrase: String): String {
        val info = keystore?.generateKey(passphrase)
        return info?.address ?: ""
    }

    fun signWithDilithium(address: String, passphrase: String, message: ByteArray): ByteArray {
        val privKeyBytes = keystore?.loadKey(address, passphrase)
            ?: throw IllegalStateException("Key not found")
        return keystore?.sign(privKeyBytes, message) ?: ByteArray(0)
    }

    fun listDilithiumKeys(): String = keystore?.listKeys() ?: "[]"

    // MARK: - Helpers

    private fun getDataDir(): String {
        val dir = File(context.filesDir, "smartlight-node")
        dir.mkdirs()
        return dir.absolutePath
    }

    private fun parseBehaviorScore(json: String): BehaviorScoreData {
        return try {
            val obj = JSONObject(json)
            BehaviorScoreData(
                total = obj.optLong("total", 5000),
                liveness = obj.optLong("liveness", 10000),
                correctness = obj.optLong("correctness", 10000),
                cooperation = obj.optLong("cooperation", 10000),
                consistency = obj.optLong("consistency", 10000),
                signalSovereignty = obj.optLong("signalSovereignty", 5000)
            )
        } catch (_: Exception) {
            BehaviorScoreData()
        }
    }

    private fun parseRewardStats(json: String): RewardStatsData {
        return try {
            val obj = JSONObject(json)
            RewardStatsData(
                totalRewards = obj.optString("totalRewards", "0"),
                epochRewards = obj.optString("epochRewards", "0"),
                currentEpoch = obj.optLong("currentEpoch", 0)
            )
        } catch (_: Exception) {
            RewardStatsData()
        }
    }
}

data class BehaviorScoreData(
    val total: Long = 5000,
    val liveness: Long = 10000,
    val correctness: Long = 10000,
    val cooperation: Long = 10000,
    val consistency: Long = 10000,
    val signalSovereignty: Long = 5000
)

data class RewardStatsData(
    val totalRewards: String = "0",
    val epochRewards: String = "0",
    val currentEpoch: Long = 0
) {
    /** Format wei to PROBE with 4 decimals */
    val totalFormatted: String get() = formatProbe(totalRewards)
    val epochFormatted: String get() = formatProbe(epochRewards)

    private fun formatProbe(wei: String): String {
        val value = wei.toBigDecimalOrNull() ?: return "0.0000"
        val probe = value.divide(1e18.toBigDecimal(), 4, java.math.RoundingMode.HALF_UP)
        return probe.toPlainString()
    }
}
