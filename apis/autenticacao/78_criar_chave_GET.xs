// Criar chave de Login
// 
// sdKF1LRsAiFXG54Q3ABzrE2quD926BDh7UR17AkQzRTRLw50VhBeEytUxc3VT0zAu0mRDPg7zxWt7cr58ZEGEXviB2eDbmhJFxuaqUMwjy1Dz1dI2QQTzesMiaUOPZlZvcDGLjRUxMw8atZoby9z0tv6DhePy6QhHkPCZMMD7xKL71L3oiOrHSglv2uKGRpaQ8bv1Q9pam53Tf4r2P5mlTsBu93VtCCsTWBuyvZfQQ9SkjjiioLmE2uoy7ZeHdcKhBmA5bzSJHB9Rg4Q3EMfH2ra8w8zgK0OqgtuDkkWaxwLsAkwC6ftasZATGPOx69kMwHYWurqCFwjaXFvA1BqhOorF5mlDjpT493IaSAULfUtW1fBBy1oduoHX8utMfmSRwuO0KlVPvD5F9aZveaOavebXlxhcaxM5hDpESTWXc78C9Lfb6I0aIBAOU1h1cdtWOAmic6HajwDfQ1zEGs12SzSP9kRMVHMRMEwFMdbCZOp92feTqLxAjtrB0JhoslO
query CriarChave verb=GET {
  api_group = "Autenticacao"

  input {
    uuid? id? {
      table = "Usuario"
    }
  
    // chave
    text chave? filters=trim
  }

  stack {
    precondition ($input.chave == "sdKF1LRsAiFXG54Q3ABzrE2quD926BDh7UR17AkQzRTRLw50VhBeEytUxc3VT0zAu0mRDPg7zxWt7cr58ZEGEXviB2eDbmhJFxuaqUMwjy1Dz1dI2QQTzesMiaUOPZlZvcDGLjRUxMw8atZoby9z0tv6DhePy6QhHkPCZMMD7xKL71L3oiOrHSglv2uKGRpaQ8bv1Q9pam53Tf4r2P5mlTsBu93VtCCsTWBuyvZfQQ9SkjjiioLmE2uoy7ZeHdcKhBmA5bzSJHB9Rg4Q3EMfH2ra8w8zgK0OqgtuDkkWaxwLsAkwC6ftasZATGPOx69kMwHYWurqCFwjaXFvA1BqhOorF5mlDjpT493IaSAULfUtW1fBBy1oduoHX8utMfmSRwuO0KlVPvD5F9aZveaOavebXlxhcaxM5hDpESTWXc78C9Lfb6I0aIBAOU1h1cdtWOAmic6HajwDfQ1zEGs12SzSP9kRMVHMRMEwFMdbCZOp92feTqLxAjtrB0JhoslO") {
      error = "Invalid Credentials."
    }
  
    db.get Usuario {
      field_name = "id"
      field_value = $input.id
    } as $Usuario1
  
    precondition ($Usuario1 != null) {
      error = "Invalid Credentials."
    }
  
    security.create_auth_token {
      table = "Usuario"
      extras = {}
      expiration = 86400
      id = $Usuario1.id
    } as $authToken
  }

  response = $authToken
}