package com.plexus.data.api

import com.plexus.data.models.*
import retrofit2.http.*

interface PlexusApi {

    // Auth
    @POST("auth/login")
    suspend fun login(@Body body: Map<String, String>): TokenPair

    @POST("auth/register")
    suspend fun register(@Body body: Map<String, String>): TokenPair

    @POST("auth/refresh")
    suspend fun refreshToken(@Body body: Map<String, String>): TokenPair

    @POST("auth/logout")
    suspend fun logout(@Body body: Map<String, String>)

    // Me
    @GET("me")
    suspend fun getMe(): User

    // Organizations
    @GET("orgs")
    suspend fun listOrgs(): ListResponse<Organization>

    @POST("orgs")
    suspend fun createOrg(@Body body: Map<String, String>): Organization

    @GET("orgs/{orgSlug}")
    suspend fun getOrg(@Path("orgSlug") orgSlug: String): Organization

    @GET("orgs/{orgSlug}/members")
    suspend fun listOrgMembers(@Path("orgSlug") orgSlug: String): ListResponse<OrgMember>

    // Projects
    @GET("orgs/{orgSlug}/projects")
    suspend fun listProjects(@Path("orgSlug") orgSlug: String): ListResponse<Project>

    @POST("orgs/{orgSlug}/projects")
    suspend fun createProject(
        @Path("orgSlug") orgSlug: String,
        @Body body: Map<String, String>,
    ): Project

    // Statuses
    @GET("orgs/{orgSlug}/projects/{projectKey}/statuses")
    suspend fun listStatuses(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
    ): ListResponse<Status>

    @POST("orgs/{orgSlug}/projects/{projectKey}/statuses")
    suspend fun createStatus(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
        @Body body: CreateStatusBody,
    ): CreatedId

    @DELETE("orgs/{orgSlug}/projects/{projectKey}/statuses/{statusId}")
    suspend fun deleteStatus(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
        @Path("statusId") statusId: String,
    )

    @GET("orgs/{orgSlug}/projects/{projectKey}/issue-types")
    suspend fun listIssueTypes(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
    ): ListResponse<IssueType>

    // Issues
    @GET("orgs/{orgSlug}/projects/{projectKey}/issues")
    suspend fun listIssues(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
        @Query("status_id") statusId: String? = null,
        @Query("sprint_id") sprintId: String? = null,
        @Query("priority") priority: String? = null,
    ): ListResponse<Issue>

    @GET("orgs/{orgSlug}/projects/{projectKey}/issues/{number}")
    suspend fun getIssue(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
        @Path("number") number: Int,
    ): Issue

    @POST("orgs/{orgSlug}/projects/{projectKey}/issues")
    suspend fun createIssue(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
        @Body body: Map<String, @JvmSuppressWildcards Any>,
    ): Map<String, Any>

    @PATCH("orgs/{orgSlug}/projects/{projectKey}/issues/{number}")
    suspend fun updateIssue(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
        @Path("number") number: Int,
        @Body body: Map<String, @JvmSuppressWildcards Any?>,
    )

    @POST("orgs/{orgSlug}/projects/{projectKey}/issues/{number}/move")
    suspend fun moveIssue(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
        @Path("number") number: Int,
        @Body body: Map<String, @JvmSuppressWildcards Any?>,
    )

    // Comments
    @GET("orgs/{orgSlug}/projects/{projectKey}/issues/{number}/comments")
    suspend fun listComments(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
        @Path("number") number: Int,
    ): ListResponse<Comment>

    @POST("orgs/{orgSlug}/projects/{projectKey}/issues/{number}/comments")
    suspend fun createComment(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
        @Path("number") number: Int,
        @Body body: Map<String, String>,
    ): Map<String, String>

    // Sprints
    @GET("orgs/{orgSlug}/projects/{projectKey}/sprints")
    suspend fun listSprints(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
    ): ListResponse<Sprint>

    @POST("orgs/{orgSlug}/projects/{projectKey}/sprints/{sprintId}/start")
    suspend fun startSprint(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
        @Path("sprintId") sprintId: String,
    )

    @POST("orgs/{orgSlug}/projects/{projectKey}/sprints/{sprintId}/complete")
    suspend fun completeSprint(
        @Path("orgSlug") orgSlug: String,
        @Path("projectKey") projectKey: String,
        @Path("sprintId") sprintId: String,
        @Body body: Map<String, @JvmSuppressWildcards String?> = emptyMap(),
    )

    // Notifications
    @GET("notifications")
    suspend fun listNotifications(): ListResponse<Notification>

    @POST("notifications/{id}/read")
    suspend fun markNotificationRead(@Path("id") id: String)

    @POST("notifications/read-all")
    suspend fun markAllRead()
}
