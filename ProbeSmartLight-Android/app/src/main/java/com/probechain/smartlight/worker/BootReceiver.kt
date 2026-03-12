// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.worker

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent

/**
 * BootReceiver restarts the SmartLight sync worker after device reboot.
 */
class BootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action == Intent.ACTION_BOOT_COMPLETED) {
            SyncWorker.schedule(context)
        }
    }
}
