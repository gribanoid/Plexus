package com.plexus.ui.issues

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import com.plexus.data.models.OrgMember
import com.plexus.data.models.Status

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EditIssueDialog(
    initialTitle: String,
    initialStatusId: String,
    initialAssigneeId: String?,
    statuses: List<Status>,
    members: List<OrgMember>,
    isSubmitting: Boolean,
    error: String?,
    onDismiss: () -> Unit,
    onConfirm: (title: String, statusId: String, assigneeId: String?) -> Unit,
) {
    var title by remember(initialTitle) { mutableStateOf(initialTitle) }
    var selectedStatusId by remember(initialStatusId) { mutableStateOf(initialStatusId) }
    var selectedAssigneeId by remember(initialAssigneeId) { mutableStateOf(initialAssigneeId ?: "") }
    var statusExpanded by remember { mutableStateOf(false) }
    var assigneeExpanded by remember { mutableStateOf(false) }

    val selectedStatus = statuses.firstOrNull { it.id == selectedStatusId }
    val selectedAssigneeLabel = members.firstOrNull { it.id == selectedAssigneeId }?.displayName ?: "Unassigned"

    Dialog(onDismissRequest = onDismiss) {
        Card {
            Column(
                modifier = Modifier.padding(24.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Text("Edit issue", style = MaterialTheme.typography.titleLarge)

                OutlinedTextField(
                    value = title,
                    onValueChange = { title = it },
                    label = { Text("Title") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                )

                ExposedDropdownMenuBox(
                    expanded = statusExpanded,
                    onExpandedChange = { statusExpanded = it },
                ) {
                    OutlinedTextField(
                        value = selectedStatus?.name ?: "Select status",
                        onValueChange = {},
                        readOnly = true,
                        label = { Text("Status") },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = statusExpanded) },
                        modifier = Modifier.menuAnchor().fillMaxWidth(),
                    )
                    ExposedDropdownMenu(
                        expanded = statusExpanded,
                        onDismissRequest = { statusExpanded = false },
                    ) {
                        statuses.forEach { status ->
                            DropdownMenuItem(
                                text = { Text(status.name) },
                                onClick = {
                                    selectedStatusId = status.id
                                    statusExpanded = false
                                },
                            )
                        }
                    }
                }

                ExposedDropdownMenuBox(
                    expanded = assigneeExpanded,
                    onExpandedChange = { assigneeExpanded = it },
                ) {
                    OutlinedTextField(
                        value = selectedAssigneeLabel,
                        onValueChange = {},
                        readOnly = true,
                        label = { Text("Assignee") },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = assigneeExpanded) },
                        modifier = Modifier.menuAnchor().fillMaxWidth(),
                    )
                    ExposedDropdownMenu(
                        expanded = assigneeExpanded,
                        onDismissRequest = { assigneeExpanded = false },
                    ) {
                        DropdownMenuItem(
                            text = { Text("Unassigned") },
                            onClick = {
                                selectedAssigneeId = ""
                                assigneeExpanded = false
                            },
                        )
                        members.forEach { member ->
                            DropdownMenuItem(
                                text = { Text(member.displayName) },
                                onClick = {
                                    selectedAssigneeId = member.id
                                    assigneeExpanded = false
                                },
                            )
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
                            onConfirm(
                                title.trim(),
                                selectedStatusId,
                                selectedAssigneeId.ifBlank { null },
                            )
                        },
                        enabled = title.trim().isNotEmpty() && selectedStatusId.isNotBlank() && !isSubmitting,
                    ) {
                        Text(if (isSubmitting) "Saving…" else "Save")
                    }
                }
            }
        }
    }
}
