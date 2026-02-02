# GPG Key Setup Guide for Terraform Provider Publishing

This guide explains how to create and configure GPG keys for signing Terraform provider releases.

## Why GPG Signing?

The Terraform Registry requires all provider releases to be signed with a GPG key. This ensures:
- **Authenticity**: Users can verify the provider comes from you
- **Integrity**: Users can verify the provider hasn't been tampered with

## Step 1: Generate a GPG Key

### Option A: Generate a new key (recommended)

```bash
# Generate a new GPG key pair
gpg --full-generate-key
```

When prompted:
1. **Key type**: Select `(1) RSA and RSA` (default)
2. **Key size**: Enter `4096` for maximum security
3. **Expiration**: Choose based on your needs (0 = never expires, or set a date)
4. **Real name**: Enter your name or organization name
5. **Email**: Enter your email address
6. **Comment**: Optional, can leave blank
7. **Passphrase**: Create a strong passphrase (you'll need this for GitHub secrets)

### Option B: Use an existing key

If you already have a GPG key, you can use it. List your keys:

```bash
gpg --list-secret-keys --keyid-format=long
```

## Step 2: Get Your Key ID and Fingerprint

List your keys to find the key ID:

```bash
gpg --list-secret-keys --keyid-format=long
```

Output example:
```
/Users/you/.gnupg/pubring.kbx
-----------------------------
sec   rsa4096/ABCD1234EFGH5678 2024-01-01 [SC]
      FINGERPRINT1234567890ABCDEF1234567890ABCDEF
uid                 [ultimate] Your Name <your.email@example.com>
ssb   rsa4096/IJKL9012MNOP3456 2024-01-01 [E]
```

In this example:
- **Key ID**: `ABCD1234EFGH5678` (the part after `rsa4096/`)
- **Fingerprint**: `FINGERPRINT1234567890ABCDEF1234567890ABCDEF`

## Step 3: Export Your Private Key

Export the private key in ASCII armor format:

```bash
# Replace ABCD1234EFGH5678 with your actual key ID
gpg --armor --export-secret-keys ABCD1234EFGH5678 > private-key.asc
```

View the contents (you'll need this for GitHub):

```bash
cat private-key.asc
```

The output will look like:
```
-----BEGIN PGP PRIVATE KEY BLOCK-----

lQdGBGV...
...many lines of base64...
...
-----END PGP PRIVATE KEY BLOCK-----
```

## Step 4: Export Your Public Key

Export the public key (needed for Terraform Registry):

```bash
# Replace ABCD1234EFGH5678 with your actual key ID
gpg --armor --export ABCD1234EFGH5678 > public-key.asc
```

## Step 5: Add Secrets to GitHub

1. Go to your GitHub repository
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**

Add these secrets:

### GPG_PRIVATE_KEY
- **Name**: `GPG_PRIVATE_KEY`
- **Value**: Paste the entire contents of `private-key.asc` (including the BEGIN and END lines)

### GPG_PASSPHRASE
- **Name**: `GPG_PASSPHRASE`
- **Value**: The passphrase you created when generating the key

## Step 6: Register with Terraform Registry

1. Go to [registry.terraform.io](https://registry.terraform.io)
2. Sign in with your GitHub account
3. Click **Publish** → **Provider**
4. Select your repository
5. When prompted for the GPG public key, paste the contents of `public-key.asc`

## Step 7: Clean Up

After adding secrets to GitHub, securely delete the exported key files:

```bash
# Securely delete the private key file
rm -P private-key.asc  # macOS
# or
shred -u private-key.asc  # Linux

# You can keep the public key or delete it
rm public-key.asc
```

## Verifying Your Setup

### Test the key locally

```bash
# Create a test file
echo "test" > test.txt

# Sign it
gpg --detach-sign --armor test.txt

# Verify the signature
gpg --verify test.txt.asc test.txt

# Clean up
rm test.txt test.txt.asc
```

### Test the GitHub workflow

1. Create a test tag:
   ```bash
   git tag v0.0.1-test
   git push origin v0.0.1-test
   ```

2. Check the Actions tab in GitHub to see if the release workflow succeeds

3. Delete the test tag if needed:
   ```bash
   git tag -d v0.0.1-test
   git push origin :refs/tags/v0.0.1-test
   ```

## Common Issues

### "gpg: signing failed: No secret key"
- Ensure the key ID matches your actual key
- Check that the private key was exported correctly

### "gpg: decryption failed: No secret key"
- The passphrase might be incorrect
- The key might not be properly imported

### GitHub Action fails with GPG error
- Verify `GPG_PRIVATE_KEY` secret contains the full key including headers
- Verify `GPG_PASSPHRASE` is correct
- Check that the key hasn't expired

## Key Management Best Practices

1. **Backup your key**: Store a backup of your private key in a secure location
2. **Use a strong passphrase**: At least 16 characters with mixed case, numbers, and symbols
3. **Set an expiration date**: Consider setting keys to expire in 2-3 years
4. **Revoke compromised keys**: If your key is compromised, revoke it immediately

## Quick Reference Commands

```bash
# List all keys
gpg --list-keys

# List secret keys
gpg --list-secret-keys --keyid-format=long

# Generate new key
gpg --full-generate-key

# Export public key
gpg --armor --export KEY_ID > public.asc

# Export private key
gpg --armor --export-secret-keys KEY_ID > private.asc

# Import a key
gpg --import key.asc

# Delete a key
gpg --delete-secret-keys KEY_ID
gpg --delete-keys KEY_ID
```
