INSERT INTO summary
(uuid, url, description, start, end, total_time, requests_count, success_req, fail_req, data_transferred, req_per_sec, avg_req_time, min_req_time, max_req_time, p50_req_time, p75_req_time, p90_req_time, p99_req_time, std_deviation, loader_uuid)
VALUES(:uuid, :url, :description, :start, :end, :total_time, :requests_count, :success_req, :fail_req, :data_transferred, :req_per_sec, :avg_req_time, :min_req_time, :max_req_time, :p50_req_time, :p75_req_time, :p90_req_time, :p99_req_time, :std_deviation, :loader_uuid)
RETURNING uuid;