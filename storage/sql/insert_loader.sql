INSERT INTO loader
(uuid, url, name, description, aggregate_window, gather_full_requests_stats, gather_aggregate_requests_stats, method, http_engine, skip_verify, ca, cert, key, benchmark_timeout, body)
VALUES (:uuid, :url, :name, :description, :aggregate_window, :gather_full_requests_stats, :gather_aggregate_requests_stats, :method, :http_engine, :skip_verify, :ca, :cert, :key, :benchmark_timeout, :body)
RETURNING uuid;