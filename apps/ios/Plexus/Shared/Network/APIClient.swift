import Foundation

/// Central HTTP client for all Plexus API calls.
/// Generated models live in Network/Generated/ (produced by openapi-generator).
final class APIClient {
    static let shared = APIClient()

    private let baseURL: URL
    private let session: URLSession
    private let keychain = KeychainStore.shared

    private init() {
        let urlString = Bundle.main.object(forInfoDictionaryKey: "PLEXUS_API_URL") as? String
            ?? "http://localhost:8080/api/v1"
        baseURL = URL(string: urlString)!
        session = URLSession(configuration: .default)
    }

    // MARK: - Generic request

    func request<T: Decodable>(
        _ endpoint: String,
        method: String = "GET",
        body: (any Encodable)? = nil
    ) async throws -> T {
        var url = baseURL.appendingPathComponent(endpoint)
        _ = url  // suppress unused-var warning; used below

        var request = URLRequest(url: baseURL.appendingPathComponent(endpoint))
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let token = keychain.accessToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        if let body {
            request.httpBody = try JSONEncoder.iso8601.encode(body)
        }

        let (data, response) = try await session.data(for: request)

        guard let http = response as? HTTPURLResponse else {
            throw APIError.invalidResponse
        }

        if http.statusCode == 401 {
            throw APIError.unauthorized
        }

        if !(200..<300).contains(http.statusCode) {
            let serverError = try? JSONDecoder().decode(ServerError.self, from: data)
            throw APIError.serverError(http.statusCode, serverError?.error ?? "Unknown error")
        }

        return try JSONDecoder.iso8601.decode(T.self, from: data)
    }

    func requestEmpty(
        _ endpoint: String,
        method: String,
        body: (any Encodable)? = nil
    ) async throws {
        var request = URLRequest(url: baseURL.appendingPathComponent(endpoint))
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let token = keychain.accessToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        if let body {
            request.httpBody = try JSONEncoder.iso8601.encode(body)
        }

        let (_, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse,
              (200..<300).contains(http.statusCode) else {
            throw APIError.invalidResponse
        }
    }
}

// MARK: - Errors

enum APIError: LocalizedError {
    case invalidResponse
    case unauthorized
    case serverError(Int, String)

    var errorDescription: String? {
        switch self {
        case .invalidResponse: return "Invalid server response"
        case .unauthorized: return "Session expired. Please sign in again."
        case .serverError(_, let msg): return msg
        }
    }
}

struct ServerError: Decodable {
    let error: String
}

// MARK: - Encoder/Decoder helpers

extension JSONEncoder {
    static let iso8601: JSONEncoder = {
        let enc = JSONEncoder()
        enc.dateEncodingStrategy = .iso8601
        enc.keyEncodingStrategy = .convertToSnakeCase
        return enc
    }()
}

extension JSONDecoder {
    static let iso8601: JSONDecoder = {
        let dec = JSONDecoder()
        dec.dateDecodingStrategy = .iso8601
        dec.keyDecodingStrategy = .convertFromSnakeCase
        return dec
    }()
}
