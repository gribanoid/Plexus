package com.plexus.ui.issues

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import com.plexus.data.models.IssueType

private val PRIORITIES = listOf(
    "urgent" to "Urgent",
    "high" to "High",
    "medium" to "Medium",
    "low" to "Low",
    "no_priority" to "No priority",
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CreateIssueDialog(
    onDismiss: () -> Unit,
    onConfirm: (title: String, typeId: String, priority: String) -> Unit,
    issueTypes: List<IssueType>,
    isLoadingTypes: Boolean,
    isSubmitting: Boolean,
    error: String?,
) {
    var title by remember { mutableStateOf("") }
    var selectedTypeId by remember { mutableStateOf<String?>(null) }
    var selectedPriority by remember { mutableStateOf("medium") }
    var typeExpanded by remember { mutableStateOf(false) }
    var priorityExpanded by remember { mutableStateOf(false) }

    LaunchedEffect(issueTypes) {
        if (selectedTypeId == null && issueTypes.isNotEmpty()) {
            selectedTypeId = issueTypes.first().id
        }
    }

    val selectedType = issueTypes.firstOrNull { it.id == selectedTypeId }
    val selectedPriorityLabel = PRIORITIES.firstOrNull { it.first == selectedPriority }?.second ?: "Medium"

    Dialog(onDismissRequest = onDismiss) {
        Card {
            Column(
                modifier = Modifier.padding(24.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Text("Create issue", style = MaterialTheme.typography.titleLarge)

                OutlinedTextField(
                    value = title,
                    onValueChange = { title = it },
                    label = { Text("Title") },
                    placeholder = { Text("What needs to be done?") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                )

                if (isLoadingTypes) {
                    LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
                } else {
                    ExposedDropdownMenuBox(
                        expanded = typeExpanded,
                        onExpandedChange = { typeExpanded = it },
                    ) {
                        OutlinedTextField(
                            value = selectedType?.name ?: "Select type",
                            onValueChange = {},
                            readOnly = true,
                            label = { Text("Type") },
                            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = typeExpanded) },
                            modifier = Modifier.menuAnchor().fillMaxWidth(),
                        )
                        ExposedDropdownMenu(
                            expanded = typeExpanded,
                            onDismissRequest = { typeExpanded = false },
                        ) {
                            issueTypes.forEach { type ->
                                DropdownMenuItem(
                                    text = { Text(type.name) },
                                    onClick = {
                                        selectedTypeId = type.id
                                        typeExpanded = false
                                    },
                                )
                            }
                        }
                    }

                    ExposedDropdownMenuBox(
                        expanded = priorityExpanded,
                        onExpandedChange = { priorityExpanded = it },
                    ) {
                        OutlinedTextField(
                            value = selectedPriorityLabel,
                            onValueChange = {},
                            readOnly = true,
                            label = { Text("Priority") },
                            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = priorityExpanded) },
                            modifier = Modifier.menuAnchor().fillMaxWidth(),
                        )
                        ExposedDropdownMenu(
                            expanded = priorityExpanded,
                            onDismissRequest = { priorityExpanded = false },
                        ) {
                            PRIORITIES.forEach { (value, label) ->
                                DropdownMenuItem(
                                    text = { Text(label) },
                                    onClick = {
                                        selectedPriority = value
                                        priorityExpanded = false
                                    },
                                )
                            }
                        }
                    }
                }

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
                            val typeId = selectedTypeId ?: return@Button
                            onConfirm(title.trim(), typeId, selectedPriority)
                        },
                        enabled = title.trim().isNotEmpty() &&
                            selectedTypeId != null &&
                            !isLoadingTypes &&
                            !isSubmitting,
                    ) {
                        Text(if (isSubmitting) "Creating…" else "Create")
                    }
                }
            }
        }
    }
}
