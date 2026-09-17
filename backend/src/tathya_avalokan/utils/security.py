import base64
import hashlib
import json
import logging
import os
from typing import Any

from cryptography.fernet import Fernet

logger = logging.getLogger(__name__)


# Master encryption key derivation
def _get_fernet_cipher() -> Fernet:
    secret = os.getenv("TATHYA_ENCRYPTION_KEY")
    if not secret:
        logger.warning(
            "⚠️  TATHYA_ENCRYPTION_KEY is not set. "
            "Falling back to a deterministic DEVELOPMENT key. "
            "This is INSECURE and must never be used in production. "
            "Generate a key with: "
            "python -c \"from cryptography.fernet import Fernet; print(Fernet.generate_key().decode())\""
        )
        # Derive a deterministic development key from a known fallback passphrase
        fallback = os.getenv("APP_SECRET_KEY", "tathya-avalokan-default-dev-secret-key-32b")
        digest = hashlib.sha256(fallback.encode("utf-8")).digest()
        key = base64.urlsafe_b64encode(digest)
    else:
        # Use the provided key directly (must be 44-char URL-safe base64) or derive it
        if len(secret) == 44:
            key = secret.encode("utf-8")
        else:
            digest = hashlib.sha256(secret.encode("utf-8")).digest()
            key = base64.urlsafe_b64encode(digest)
        logger.debug("Fernet cipher initialized from TATHYA_ENCRYPTION_KEY.")
    return Fernet(key)


def encrypt_credentials(credentials: dict[str, Any]) -> str:
    """Symmetrically encrypts sensitive credentials dictionary to Fernet ciphertext string."""
    cipher = _get_fernet_cipher()
    raw_bytes = json.dumps(credentials).encode("utf-8")
    return cipher.encrypt(raw_bytes).decode("utf-8")


def decrypt_credentials(encrypted_text: str) -> dict[str, Any]:
    """Decrypts Fernet ciphertext string back to raw credentials dictionary."""
    cipher = _get_fernet_cipher()
    decrypted_bytes = cipher.decrypt(encrypted_text.encode("utf-8"))
    return json.loads(decrypted_bytes.decode("utf-8"))  # type: ignore[no-any-return]


def mask_connection_uri(uri: str) -> str:
    """Masks sensitive password credentials inside a standard connection URI."""
    if "@" in uri and "://" in uri:
        protocol_part, rest = uri.split("://", 1)
        auth_part, host_part = rest.split("@", 1)
        if ":" in auth_part:
            user = auth_part.split(":", 1)[0]
            return f"{protocol_part}://{user}:********@{host_part}"
    return uri
