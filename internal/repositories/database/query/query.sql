-- name: RegisterUser :exec
INSERT INTO users (login, password)
VALUES ($1, $2);

-- name: GetUser :one
SELECT login, password, current, withdrawn
FROM users
WHERE login = $1;

-- name: UploadOrder :one
INSERT INTO orders (number, user_login)
VALUES ($1, $2)
RETURNING user_login;

-- name: UploadWithdrawal :exec
INSERT INTO withdrawals (number, user_login, sum)
VALUES ($1, $2, $3);

-- name: GetUserOrders :many
SELECT number, uploaded_at, user_login, status, accrual
FROM orders
WHERE user_login = $1
ORDER BY uploaded_at DESC;

-- name: GetUserWithdrawals :many
SELECT number, processed_at, user_login, sum
FROM withdrawals
WHERE user_login = $1
ORDER BY processed_at DESC;

-- name: UpdateUserBalance :exec
UPDATE users
SET current = current + $2, withdrawn = withdrawn + $3
WHERE login = $1;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2
WHERE number = $1;

-- name: UpdateOrderAccrual :exec
UPDATE orders
SET accrual = $2
WHERE number = $1;

-- name: GetOrdersWithStatus :many
SELECT number, uploaded_at, user_login, status, accrual
FROM orders
WHERE status = $1;

-- name: GetUnprocessedOrders :many
SELECT number, uploaded_at, user_login, status, accrual
FROM orders
WHERE status in ('NEW', 'PROCESSING');
