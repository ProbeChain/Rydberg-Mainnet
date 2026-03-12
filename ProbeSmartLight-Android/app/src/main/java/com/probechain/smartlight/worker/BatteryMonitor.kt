// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.worker

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.os.BatteryManager

/**
 * BatteryMonitor watches battery state changes and switches SmartLight power mode:
 * - Charging → Full (ACK + GNSS + Agent + Heartbeat)
 * - Battery > 30% → Eco (ACK + Heartbeat only)
 * - Battery < 15% → Sleep (Sync only)
 */
class BatteryMonitor(private val onPowerModeChange: (Int) -> Unit) : BroadcastReceiver() {

    fun register(context: Context) {
        val filter = IntentFilter().apply {
            addAction(Intent.ACTION_BATTERY_CHANGED)
            addAction(Intent.ACTION_POWER_CONNECTED)
            addAction(Intent.ACTION_POWER_DISCONNECTED)
        }
        context.registerReceiver(this, filter)
    }

    fun unregister(context: Context) {
        context.unregisterReceiver(this)
    }

    override fun onReceive(context: Context, intent: Intent) {
        val status = intent.getIntExtra(BatteryManager.EXTRA_STATUS, -1)
        val level = intent.getIntExtra(BatteryManager.EXTRA_LEVEL, -1)
        val scale = intent.getIntExtra(BatteryManager.EXTRA_SCALE, 100)
        val batteryPct = level * 100 / scale

        val isCharging = status == BatteryManager.BATTERY_STATUS_CHARGING ||
                         status == BatteryManager.BATTERY_STATUS_FULL

        val mode = when {
            isCharging -> 0 // Full
            batteryPct > 30 -> 1 // Eco
            else -> 2 // Sleep
        }

        onPowerModeChange(mode)
    }
}
