package visdata

// Match de setor alinhado ao terminal (particao 2 digitos, zonauser 3 digitos).
const cameraSetorWhereSQL = `
WHERE c.id_dispositivo = $1
  AND LPAD(NULLIF(REGEXP_REPLACE(TRIM(COALESCE(c.particao, '')), '[^0-9]', '', 'g'), ''), 2, '0')
    = LPAD(NULLIF(REGEXP_REPLACE(TRIM(COALESCE($2, '')), '[^0-9]', '', 'g'), ''), 2, '0')
  AND (
    TRIM(COALESCE(c.zonauser, '')) = TRIM(COALESCE($3, ''))
    OR (
      NULLIF(REGEXP_REPLACE(TRIM(COALESCE(c.zonauser, '')), '[^0-9]', '', 'g'), '') IS NOT NULL
      AND LPAD(NULLIF(REGEXP_REPLACE(TRIM(COALESCE(c.zonauser, '')), '[^0-9]', '', 'g'), ''), 3, '0')
        = LPAD(NULLIF(REGEXP_REPLACE(TRIM(COALESCE($3, '')), '[^0-9]', '', 'g'), ''), 3, '0')
    )
  )`

func prepareSetorQueryParams(particao, zonauser string) (string, string) {
	return padDigits(particao, 2), normalizeZonaUser(zonauser)
}
