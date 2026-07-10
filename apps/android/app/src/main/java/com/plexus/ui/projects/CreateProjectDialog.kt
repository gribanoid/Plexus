package com.plexus.ui.projects

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import androidx.compose.foundation.text.KeyboardOptions

private fun suggestProjectKey(name: String): String {
    val words = name.trim().split(Regex("\\s+"))
    val key = buildString {
        for (word in words) {
            if (length >= 4) break
            word.firstOrNull { it.isLetterOrDigit() }?.uppercaseChar()?.let { append(it) }
        }
    }
    return key.ifBlank { "PRJ" }
}

@Composable
fun CreateProjectDialog(
    orgName: String?,
    onDismiss: () -> Unit,
    onConfirm: (name: String, key: String?, description: String?) -> Unit,
    isSubmitting: Boolean,
    error: String?,
) {
    var name by remember { mutableStateOf("") }
    var key by remember { mutableStateOf("") }
    var description by remember { mutableStateOf("") }
    var keyEdited by remember { mutableStateOf(false) }

    LaunchedEffect(name) {
        if (!keyEdited && name.length >= 2) {
            key = suggestProjectKey(name)
        }
    }

    Dialog(onDismissRequest = onDismiss) {
        Card {
            Column(
                modifier = Modifier.padding(24.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Text(
                    if (orgName != null) "Create project in $orgName" else "Create project",
                    style = MaterialTheme.typography.titleLarge,
                )

                OutlinedTextField(
                    value = name,
                    onValueChange = { name = it },
                    label = { Text("Project name") },
                    placeholder = { Text("Marketing Board") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                )

                OutlinedTextField(
                    value = key,
                    onValueChange = {
                        key = it.uppercase()
                        keyEdited = true
                    },
                    label = { Text("Key") },
                    placeholder = { Text("MKT") },
                    singleLine = true,
                    keyboardOptions = KeyboardOptions(capitalization = KeyboardCapitalization.Characters),
                    modifier = Modifier.fillMaxWidth(),
                )

                OutlinedTextField(
                    value = description,
                    onValueChange = { description = it },
                    label = { Text("Description (optional)") },
                    placeholder = { Text("What is this project about?") },
                    minLines = 2,
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
                        onClick = {
                            onConfirm(
                                name.trim(),
                                key.trim().ifBlank { null },
                                description.trim().ifBlank { null },
                            )
                        },
                        enabled = name.trim().length >= 2 && !isSubmitting,
                    ) {
                        Text(if (isSubmitting) "Creating…" else "Create")
                    }
                }
            }
        }
    }
}
