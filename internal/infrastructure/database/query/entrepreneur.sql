-- name: CreateEntrepreneur :one
INSERT INTO entrepreneurs (
    inn_id, legal_name, registration_authority, registration_date,
    registration_number, legal_form, ifut_code_id, dbibt_code,
    activity_status, charter_fund, founders, email,
    phone, mhobt_code, address, director_name, sqb_api_error
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING id, inn_id, legal_name, registration_authority, registration_date,
    registration_number, legal_form, ifut_code_id, dbibt_code,
    activity_status, charter_fund, founders, email,
    phone, mhobt_code, address, director_name, sqb_api_error, created_at;

-- name: GetEntrepreneurByID :one
SELECT e.id, e.inn_id, i.name as inn_name, e.legal_name, e.registration_authority,
    e.registration_date, e.registration_number, e.legal_form, e.ifut_code_id,
    ic.name as ifut_code_name, e.dbibt_code, e.activity_status, e.charter_fund,
    e.founders, e.email, e.phone, e.mhobt_code, e.address, e.director_name,
    e.sqb_api_error, e.created_at
FROM entrepreneurs e
JOIN inns i ON e.inn_id = i.id
LEFT JOIN ifut_codes ic ON e.ifut_code_id = ic.id
WHERE e.id = $1;

-- name: UpdateEntrepreneur :one
UPDATE entrepreneurs
SET legal_name = $2,
    registration_authority = $3,
    registration_date = $4,
    registration_number = $5,
    legal_form = $6,
    ifut_code_id = $7,
    dbibt_code = $8,
    activity_status = $9,
    charter_fund = $10,
    founders = $11,
    email = $12,
    phone = $13,
    mhobt_code = $14,
    address = $15,
    director_name = $16
WHERE id = $1
RETURNING id, inn_id, legal_name, registration_authority, registration_date,
    registration_number, legal_form, ifut_code_id, dbibt_code,
    activity_status, charter_fund, founders, email,
    phone, mhobt_code, address, director_name, sqb_api_error, created_at;

-- name: DeleteEntrepreneur :exec
DELETE FROM entrepreneurs
WHERE id = $1;

-- name: SetSqbApiError :exec
UPDATE entrepreneurs
SET sqb_api_error = $2
WHERE id = $1;
