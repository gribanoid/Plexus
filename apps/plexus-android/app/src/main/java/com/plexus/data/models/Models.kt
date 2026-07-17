package com.plexus.data.models

import com.squareup.moshi.Json
import com.squareup.moshi.JsonClass

@JsonClass(generateAdapter = true)
data class TokenPair(
    @Json(name = "access_token") val accessToken: String,
    @Json(name = "refresh_token") val refreshToken: String,
    @Json(name = "expires_in") val expiresIn: Int,
)

@JsonClass(generateAdapter = true)
data class User(
    val id: String,
    val email: String,
    @Json(name = "display_name") val displayName: String,
    @Json(name = "avatar_url") val avatarUrl: String?,
    val role: String,
    @Json(name = "created_at") val createdAt: String,
)

@JsonClass(generateAdapter = true)
data class Organization(
    val id: String,
    val slug: String,
    val name: String,
    @Json(name = "logo_url") val logoUrl: String?,
    val plan: String,
    @Json(name = "my_role") val myRole: String?,
)

@JsonClass(generateAdapter = true)
data class Project(
    val id: String,
    @Json(name = "org_id") val orgId: String? = null,
    val key: String,
    val name: String,
    val description: String?,
    @Json(name = "icon_url") val iconUrl: String?,
    @Json(name = "lead_id") val leadId: String?,
)

@JsonClass(generateAdapter = true)
data class IssueType(
    val id: String,
    val name: String,
    val color: String,
    @Json(name = "icon_url") val iconUrl: String?,
)

@JsonClass(generateAdapter = true)
data class Status(
    val id: String,
    val name: String,
    val color: String,
    val category: String,
    val position: Int,
)

@JsonClass(generateAdapter = true)
data class CreateStatusBody(
    val name: String,
    val color: String,
    val category: String,
)

@JsonClass(generateAdapter = true)
data class CreatedId(
    val id: String,
)

@JsonClass(generateAdapter = true)
data class Issue(
    val id: String,
    val number: Int,
    val title: String,
    val description: String?,
    val priority: String,
    @Json(name = "status_id") val statusId: String,
    @Json(name = "type_id") val typeId: String,
    @Json(name = "assignee_id") val assigneeId: String?,
    @Json(name = "reporter_id") val reporterId: String,
    @Json(name = "sprint_id") val sprintId: String?,
    @Json(name = "story_points") val storyPoints: Double?,
    @Json(name = "due_date") val dueDate: String?,
    val position: Double,
    @Json(name = "created_at") val createdAt: String,
    @Json(name = "updated_at") val updatedAt: String,
)

@JsonClass(generateAdapter = true)
data class Comment(
    val id: String,
    val body: String,
    @Json(name = "author_id") val authorId: String,
    @Json(name = "created_at") val createdAt: String,
)

@JsonClass(generateAdapter = true)
data class Sprint(
    val id: String,
    val name: String,
    val goal: String?,
    val state: String,
    @Json(name = "start_date") val startDate: String?,
    @Json(name = "end_date") val endDate: String?,
)

@JsonClass(generateAdapter = true)
data class OrgMember(
    val id: String,
    @Json(name = "display_name") val displayName: String,
    val email: String,
    @Json(name = "avatar_url") val avatarUrl: String?,
    val role: String,
)

@JsonClass(generateAdapter = true)
data class Notification(
    val id: String,
    val type: String,
    val title: String,
    val body: String?,
    val read: Boolean,
    @Json(name = "issue_id") val issueId: String?,
    @Json(name = "created_at") val createdAt: String,
)

@JsonClass(generateAdapter = true)
data class ListResponse<T>(val items: List<T>)
