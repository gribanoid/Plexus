package com.plexus.ui.orgs

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog

@Composable
fun CreateOrgDialog(
    onDismiss: () -> Unit,
    onConfirm: (name: String, slug: String?) -> Unit,
    isSubmitting: Boolean,
    error: String?,
) {
    var name by remember { mutableStateOf("") }
    var slug by remember { mutableStateOf("") }

    Dialog(onDismissRequest = onDismiss) {
        Card {
            Column(
                modifier = Modifier.padding(24.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Text("Create workspace", style = MaterialTheme.typography.titleLarge)

                OutlinedTextField(
                    value = name,
                    onValueChange = { name = it },
                    label = { Text("Workspace name") },
                    placeholder = { Text("Acme Inc") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                )

                OutlinedTextField(
                    value = slug,
                    onValueChange = { slug = it },
                    label = { Text("URL slug (optional)") },
                    placeholder = {
                        Text(name.lowercase().replace(Regex("\\s+"), "-").ifBlank { "acme-inc" })
                    },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                )

                error?.let {
                    Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
                }

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.End,
                ) {
                    TextButton(onClick = onDismiss, enabled = !isSubmitting) {
                        Text("Cancel")
                    }
                    Spacer(Modifier.width(8.dp))
                    Button(
                        onClick = { onConfirm(name.trim(), slug.trim().ifBlank { null }) },
                        enabled = name.trim().length >= 2 && !isSubmitting,
                    ) {
                        Text(if (isSubmitting) "Creating…" else "Create")
                    }
                }
            }
        }
    }
}
