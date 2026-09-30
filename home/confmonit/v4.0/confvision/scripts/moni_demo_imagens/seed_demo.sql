-- Demo Moni — franqueado 0, eventos 990001-990003
-- Rodar no Postgres Core4: psql "$POSTGRES_URL" -f seed_demo.sql

INSERT INTO vis_evento (
    id, id_franqueado, id_cliente, conta, particao, canal,
    tipo_deteccao, snapshot_url,
    codigo_imagem_publico, imagem_liberada_em
) VALUES
(
    990001, '0', '0', '0000', '00', '001', 'movimento',
    'https://usc1.contabostorage.com/d845a20b9a864ab68dbf12c9fd773342:confvision/confvision/eventos/0/990001/snapshot.jpg',
    'yw4zz83ag4x9', NOW()
),
(
    990002, '0', '0', '0000', '00', '001', 'movimento',
    'https://usc1.contabostorage.com/d845a20b9a864ab68dbf12c9fd773342:confvision/confvision/eventos/0/990002/snapshot.jpg',
    '6rorjlqej4wy', NOW()
),
(
    990003, '0', '0', '0000', '00', '001', 'movimento',
    'https://usc1.contabostorage.com/d845a20b9a864ab68dbf12c9fd773342:confvision/confvision/eventos/0/990003/snapshot.jpg',
    'qa5m9dx7p4el', NOW()
)
ON CONFLICT (id) DO UPDATE SET
    snapshot_url = EXCLUDED.snapshot_url,
    codigo_imagem_publico = EXCLUDED.codigo_imagem_publico,
    imagem_liberada_em = NOW();
