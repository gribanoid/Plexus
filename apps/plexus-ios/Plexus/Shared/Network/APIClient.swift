import Foundation

/// Central HTTP client for all Plexus API calls.
final class APIClient {
    static let shared = APIClient()

    private let baseURL: URL
    private let session: URLSession
    private let keychain = KeychainStore.shared

    private init() {
        let urlString = Bundle.main.object(forInfoDictionaryKey: "PLEXUS_API_URL") as? String
            ?? "http://127.0.0.1:8080/api/v1"
        baseURL = URL(string: urlString)!
        session = URLSession(configuration: .default)
    }

    // MARK: - Generic request

    func request<T: Decodable>(
        _ endpoint: String,
        method: String = "GET",
        body: (any Encodable)? = nil
    ) async throws -> T {
        let (data, _) = try await perform(endpoint, method: method, body: body)
        return try JSONDecoder.iso8601.decode(T.self, from: data)
    }

    func requestEmpty(
        _ endpoint: String,
        method: String,
        body: (any Encodable)? = nil
    ) async throws {
        _ = try await perform(endpoint, method: method, body: body)
    }

    private func perform(
        _ endpoint: String,
        method: String,
        body: (any Encodable)?
    ) async throws -> (Data, HTTPURLResponse) {
        let url = makeURL(endpoint)
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let token = keychain.accessToken {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }

        if let body {
            request.httpBody = try JSONEncoder.iso8601.encode(body)
        }

        let data: Data
        let response: URLResponse
        do {
            (data, response) = try await session.data(for: request)
        } catch {
            throw APIError.network(error.localizedDescription)
        }

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

        return (data, http)
    }

    /// Join base + endpoint without percent-encoding `/` inside the path.
    private func makeURL(_ endpoint: String) -> URL {
        let base = baseURL.absoluteString.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        let path = endpoint.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        guard let url = URL(string: "\(base)/\(path)") else {
            preconditionFailure("Invalid API URL: \(base)/\(path)")
        }
        return url
    }
}

// MARK: - Errors

enum APIError: LocalizedError {
    case invalidResponse
    case unauthorized
    case network(String)
    case serverError(Int, String)

    var errorDescription: String? {
        switch self {
        case .invalidResponse: return "Invalid server response"
        case .unauthorized: return "Session expired. Please sign in again."
        case .network: return "Unable to connect to the server. Please try again."
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
        dec.keyDecodingStrategy = .convertFromSnakeCase
        // Go encodes time.Time as RFC3339 with optional fractional seconds.
        dec.dateDecodingStrategy = .custom { decoder in
            let container = try decoder.singleValueContainer()
            let raw = try container.decode(String.self)
            if let date = JSONDecoder.parseISO8601(raw) {
                return date
            }
            throw DecodingError.dataCorruptedError(
                in: container,
                debugDescription: "Invalid date: \(raw)"
            )
        }
        return dec
    }()

    /// Parse Go/RFC3339 timestamps without capturing non-Sendable formatters in the decode closure.
    private static func parseISO8601(_ raw: String) -> Date? {
        let withFraction = ISO8601DateFormatter()
        withFraction.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        if let date = withFraction.date(from: raw) { return date }
        let plain = ISO8601DateFormatter()
        plain.formatOptions = [.withInternetDateTime]
        return plain.date(from: raw)
    }
}
