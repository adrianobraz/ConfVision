<?php
/**
 * Proxy ConfService → API Go no Proxmox.
 * Usado na KingHost (sem ProxyPass): /parceiro/*, /marketplace/*, etc.
 *
 * Configure via cs-proxy.env:
 *   CS_API_BACKEND=http://185.130.61.4:2020
 */
$backend = getenv('CS_API_BACKEND');
if (!$backend) {
    $backend = 'http://185.130.61.4:2020';
}
$envFile = __DIR__ . '/cs-proxy.env';
if (is_readable($envFile)) {
    $lines = file($envFile, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES);
    if (is_array($lines)) {
        foreach ($lines as $line) {
            $line = trim($line);
            if ($line === '' || $line[0] === '#') {
                continue;
            }
            if (strpos($line, 'CS_API_BACKEND=') === 0) {
                $backend = trim(substr($line, strlen('CS_API_BACKEND=')), " \t\"'");
            }
        }
    }
}

$uri = isset($_SERVER['REQUEST_URI']) ? $_SERVER['REQUEST_URI'] : '/';
$path = parse_url($uri, PHP_URL_PATH);
if (!$path) {
    $path = '/';
}
$query = parse_url($uri, PHP_URL_QUERY);
$target = rtrim($backend, '/') . $path;
if ($query) {
    $target .= '?' . $query;
}

$method = isset($_SERVER['REQUEST_METHOD']) ? $_SERVER['REQUEST_METHOD'] : 'GET';
$body = file_get_contents('php://input');

$headers = array();
foreach ($_SERVER as $key => $value) {
    if (strpos($key, 'HTTP_') === 0) {
        $name = str_replace(' ', '-', ucwords(strtolower(str_replace('_', ' ', substr($key, 5)))));
        $lname = strtolower($name);
        if ($lname === 'host' || $lname === 'connection' || $lname === 'content-length') {
            continue;
        }
        $headers[] = $name . ': ' . $value;
    }
}
if (!empty($_SERVER['CONTENT_TYPE'])) {
    $headers[] = 'Content-Type: ' . $_SERVER['CONTENT_TYPE'];
}

$ch = curl_init($target);
curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_HEADER, true);
curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
curl_setopt($ch, CURLOPT_TIMEOUT, 60);
curl_setopt($ch, CURLOPT_FOLLOWLOCATION, false);
if ($body !== false && $body !== '' && $method !== 'GET' && $method !== 'HEAD') {
    curl_setopt($ch, CURLOPT_POSTFIELDS, $body);
}

$raw = curl_exec($ch);
if ($raw === false) {
    http_response_code(502);
    header('Content-Type: application/json; charset=utf-8');
    echo json_encode(array(
        'ok' => false,
        'erro' => 'proxy: falha ao contatar API (' . curl_error($ch) . ')',
    ));
    curl_close($ch);
    exit;
}

$status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
$headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
curl_close($ch);

$respHeaders = substr($raw, 0, $headerSize);
$respBody = substr($raw, $headerSize);

http_response_code($status);
foreach (explode("\r\n", $respHeaders) as $line) {
    if ($line === '' || strpos($line, 'HTTP/') === 0) {
        continue;
    }
    $lower = strtolower($line);
    if (strpos($lower, 'transfer-encoding:') === 0) {
        continue;
    }
    if (strpos($lower, 'connection:') === 0) {
        continue;
    }
    header($line, false);
}
echo $respBody;
