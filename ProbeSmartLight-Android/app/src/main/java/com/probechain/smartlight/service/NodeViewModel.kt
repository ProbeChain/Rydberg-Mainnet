// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.service

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.probechain.smartlight.SmartLightApp
import com.probechain.smartlight.core.BehaviorScoreData
import com.probechain.smartlight.core.GoNodeBridge
import com.probechain.smartlight.core.RewardStatsData
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch

/**
 * NodeViewModel manages the SmartLight node state for the UI layer.
 * Exposes reactive StateFlows for Compose screens.
 */
class NodeViewModel(application: Application) : AndroidViewModel(application) {

    private val bridge = GoNodeBridge(application)

    private val _isRunning = MutableStateFlow(false)
    val isRunning: StateFlow<Boolean> = _isRunning.asStateFlow()

    private val _syncedBlock = MutableStateFlow(0L)
    val syncedBlock: StateFlow<Long> = _syncedBlock.asStateFlow()

    private val _peerCount = MutableStateFlow(0)
    val peerCount: StateFlow<Int> = _peerCount.asStateFlow()

    private val _powerModeName = MutableStateFlow("Full")
    val powerModeName: StateFlow<String> = _powerModeName.asStateFlow()

    private val _behaviorScore = MutableStateFlow(BehaviorScoreData())
    val behaviorScore: StateFlow<BehaviorScoreData> = _behaviorScore.asStateFlow()

    private val _rewardStats = MutableStateFlow(RewardStatsData())
    val rewardStats: StateFlow<RewardStatsData> = _rewardStats.asStateFlow()

    private var currentPowerMode = 0

    fun toggleNode() {
        if (_isRunning.value) {
            stopNode()
        } else {
            startNode()
        }
    }

    fun startNode() {
        try {
            bridge.start()
            _isRunning.value = true
            startPolling()
        } catch (e: Exception) {
            android.util.Log.e("NodeViewModel", "Failed to start node", e)
        }
    }

    fun stopNode() {
        bridge.stop()
        _isRunning.value = false
    }

    fun setPowerMode(mode: Int) {
        currentPowerMode = mode
        bridge.setPowerMode(mode)
        _powerModeName.value = when (mode) {
            0 -> "Full"
            1 -> "Eco"
            2 -> "Sleep"
            else -> "Unknown"
        }
    }

    fun cyclePowerMode() {
        setPowerMode((currentPowerMode + 1) % 3)
    }

    private fun startPolling() {
        // Poll node stats every 5 seconds
        viewModelScope.launch {
            while (isActive && _isRunning.value) {
                _peerCount.value = bridge.getPeerCount()
                _behaviorScore.value = bridge.getBehaviorScore()
                _rewardStats.value = bridge.getRewardStats()
                delay(5000)
            }
        }
    }

    override fun onCleared() {
        super.onCleared()
        if (_isRunning.value) {
            bridge.stop()
        }
    }
}
