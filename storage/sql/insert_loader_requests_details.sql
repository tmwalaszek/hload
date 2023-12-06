INSERT INTO loader_requests_details
(request_count, abort_after, connections, rate_limit, duration, keep_alive, request_delay, read_timeout, write_timeout, timeout, loader_uuid)
VALUES (:request_count, :abort_after, :connections, :rate_limit, :duration, :keep_alive, :request_delay, :read_timeout, :write_timeout, :timeout, :loader_uuid)