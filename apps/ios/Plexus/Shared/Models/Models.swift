import Foundation

// MARK: - Auth

struct TokenPair: Decodable {
    let accessToken: String
    let refreshToken: String
    let expiresIn: Int
}

// MARK: - User

struct User: Decodable, Identifiable {
    let id: String
    let email: String
    let displayName: String
    let avatarUrl: String?
    let role: String
    let createdAt: Date
}

// MARK: - Organization

struct Organization: Decodable, Identifiable {
    let id: String
    let slug: String
    let name: String
    let logoUrl: String?
    let plan: String
    let myRole: String?
}

// MARK: - Project

struct Project: Decodable, Identifiable {
    let id: String
    let orgId: String
    let key: String
    let name: String
    let description: String?
    let iconUrl: String?
    let leadId: String?
}

// MARK: - Status

struct Status: Decodable, Identifiable {
    let id: String
    let name: String
    let color: String
    let category: StatusCategory
    let position: Int
}

enum StatusCategory: String, Decodable {
    case todo, inProgress = "in_progress", done
}

// MARK: - Issue

struct Issue: Decodable, Identifiable {
    let id: String
    let number: Int
    let title: String
    let description: String?
    let priority: Priority
    let statusId: String
    let typeId: String
    let assigneeId: String?
    let reporterId: String
    let sprintId: String?
    let parentId: String?
    let storyPoints: Double?
    let dueDate: Date?
    let position: Double
    let createdAt: Date
    let updatedAt: Date
}

enum Priority: String, Decodable, Encodable, CaseIterable, Hashable {
    case urgent, high, medium, low
    case noPriority = "no_priority"

    var displayName: String {
        switch self {
        case .urgent: return "Urgent"
        case .high: return "High"
        case .medium: return "Medium"
        case .low: return "Low"
        case .noPriority: return "No Priority"
        }
    }

    var color: String {
        switch self {
        case .urgent: return "#EF4444"
        case .high: return "#F97316"
        case .medium: return "#EAB308"
        case .low: return "#3B82F6"
        case .noPriority: return "#9CA3AF"
        }
    }
}

// MARK: - Issue Type

struct IssueType: Decodable, Identifiable {
    let id: String
    let name: String
    let color: String
    let iconUrl: String?
}

// MARK: - Issue History

struct IssueHistory: Decodable, Identifiable {
    let id: String
    let field: String
    let oldValue: String?
    let newValue: String?
    let actorId: String
    let createdAt: Date
}

// MARK: - Comment

struct Comment: Decodable, Identifiable {
    let id: String
    let body: String
    let authorId: String
    let createdAt: Date
    let updatedAt: Date
}

// MARK: - Sprint

struct Sprint: Decodable, Identifiable {
    let id: String
    let name: String
    let goal: String?
    let state: SprintState
    let startDate: Date?
    let endDate: Date?
}

enum SprintState: String, Decodable {
    case active, closed, future
}

// MARK: - Org Member

struct OrgMember: Decodable, Identifiable {
    let id: String
    let displayName: String
    let email: String
    let avatarUrl: String?
    let role: String
}

// MARK: - Notification

struct PlexusNotification: Decodable, Identifiable {
    let id: String
    let type: String
    let title: String
    let body: String?
    let read: Bool
    let issueId: String?
    let createdAt: Date
}

// MARK: - List response wrapper

struct ListResponse<T: Decodable>: Decodable {
    let items: [T]
}
