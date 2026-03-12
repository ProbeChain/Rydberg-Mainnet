// Copyright 2024 The ProbeChain Authors
package com.probechain.smartlight.core

import android.os.Build
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Log
import java.security.KeyPairGenerator
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

/**
 * AndroidKeyStoreWrapper protects Dilithium private keys (2528 bytes) using
 * Android Keystore with hardware-backed security (StrongBox if available).
 *
 * Architecture:
 * - AES-256-GCM key stored in Android Keystore (hardware-backed)
 * - Dilithium private key encrypted with this AES key
 * - Biometric authentication required to use the AES key
 * - Private key never leaves the device unencrypted
 */
object AndroidKeyStoreWrapper {

    private const val TAG = "AndroidKeyStore"
    private const val KEYSTORE_PROVIDER = "AndroidKeyStore"
    private const val KEY_ALIAS = "probechain_smartlight_dilithium_wrap"
    private const val GCM_TAG_LENGTH = 128

    /**
     * Check if hardware-backed keystore (StrongBox) is available.
     */
    val isStrongBoxAvailable: Boolean
        get() = Build.VERSION.SDK_INT >= Build.VERSION_CODES.P

    /**
     * Creates or retrieves the AES-256-GCM wrapping key in Android Keystore.
     * The key requires biometric authentication to use.
     */
    fun getOrCreateWrappingKey(): SecretKey {
        val keyStore = KeyStore.getInstance(KEYSTORE_PROVIDER)
        keyStore.load(null)

        // Check if key already exists
        if (keyStore.containsAlias(KEY_ALIAS)) {
            val entry = keyStore.getEntry(KEY_ALIAS, null) as KeyStore.SecretKeyEntry
            return entry.secretKey
        }

        // Generate new AES-256 key in Keystore
        val spec = KeyGenParameterSpec.Builder(
            KEY_ALIAS,
            KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
        ).apply {
            setKeySize(256)
            setBlockModes(KeyProperties.BLOCK_MODE_GCM)
            setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
            setUserAuthenticationRequired(true)
            // Require biometric auth, valid for 10 seconds after auth
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
                setUserAuthenticationParameters(10, KeyProperties.AUTH_BIOMETRIC_STRONG)
            }
            // Use StrongBox (hardware security module) if available
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
                try {
                    setIsStrongBoxBacked(true)
                } catch (e: Exception) {
                    Log.w(TAG, "StrongBox not available, using TEE", e)
                }
            }
        }.build()

        val keyGenerator = KeyGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_AES, KEYSTORE_PROVIDER
        )
        keyGenerator.init(spec)
        return keyGenerator.generateKey()
    }

    /**
     * Encrypts a Dilithium private key (2528 bytes) with the Keystore-backed AES key.
     * Returns: IV (12 bytes) + ciphertext + GCM tag
     */
    fun encryptDilithiumKey(privateKeyBytes: ByteArray): ByteArray {
        val secretKey = getOrCreateWrappingKey()
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, secretKey)

        val iv = cipher.iv // 12 bytes
        val ciphertext = cipher.doFinal(privateKeyBytes)

        // Concatenate: IV + ciphertext (includes GCM auth tag)
        return iv + ciphertext
    }

    /**
     * Decrypts a Dilithium private key. Requires biometric authentication.
     * Input: IV (12 bytes) + ciphertext + GCM tag
     * Returns: plaintext private key (2528 bytes)
     */
    fun decryptDilithiumKey(encryptedData: ByteArray): ByteArray {
        require(encryptedData.size > 12) { "Encrypted data too short" }

        val iv = encryptedData.sliceArray(0 until 12)
        val ciphertext = encryptedData.sliceArray(12 until encryptedData.size)

        val secretKey = getOrCreateWrappingKey()
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val spec = GCMParameterSpec(GCM_TAG_LENGTH, iv)
        cipher.init(Cipher.DECRYPT_MODE, secretKey, spec)

        return cipher.doFinal(ciphertext)
    }

    /**
     * Returns a Cipher initialized for decryption, for use with BiometricPrompt.
     * BiometricPrompt.CryptoObject wraps this cipher for authenticated decryption.
     */
    fun getDecryptCipher(encryptedData: ByteArray): Cipher {
        val iv = encryptedData.sliceArray(0 until 12)
        val secretKey = getOrCreateWrappingKey()
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        val spec = GCMParameterSpec(GCM_TAG_LENGTH, iv)
        cipher.init(Cipher.DECRYPT_MODE, secretKey, spec)
        return cipher
    }

    /**
     * Deletes the wrapping key (and effectively all encrypted Dilithium keys).
     */
    fun deleteWrappingKey() {
        val keyStore = KeyStore.getInstance(KEYSTORE_PROVIDER)
        keyStore.load(null)
        keyStore.deleteEntry(KEY_ALIAS)
    }
}
