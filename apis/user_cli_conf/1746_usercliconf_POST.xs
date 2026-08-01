// Add UserCliConf record
query usercliconf verb=POST {
  api_group = "UserCliConf"

  input {
    email email? filters=trim|lower
    text idFranqueado? filters=trim
    text idCliente? filters=trim
    text nome? filters=trim
    text telefone1? filters=trim
    text telefone2? filters=trim
    text Descricao? filters=trim
    password senha? {
      sensitive = true
      visibility = "internal"
    }
  
    enum[] Acesso? {
      values = [
        "MEUSDADOS"
        "GRADEHORARIO"
        "MSGATENDENTE"
        "RELATORIO"
        "WHATSAPP"
        "CAMERA"
      ]
    }
  }

  stack {
    var $x1 {
      value = {}
    }
  
    conditional {
      if ($input.email != null) {
        db.add UserCliConf {
          enforce_hidden_fields = false
          data = {
            created_at  : "now"
            email       : $input.email
            senha       : $input.senha
            Data        : now
            idFranqueado: $input.idFranqueado
            idCliente   : $input.idCliente
            nome        : $input.nome
            telefone1   : $input.telefone1
            telefone2   : $input.telefone2
            Descricao   : $input.Descricao
            Acesso      : $input.Acesso
          }
        
          output = [
            "id"
            "created_at"
            "email"
            "Data"
            "idFranqueado"
            "idCliente"
            "nome"
            "telefone1"
            "telefone2"
            "Descricao"
            "Acesso"
          ]
        } as $model
      
        var.update $x1 {
          value = `$model|safe_array`
        }
      }
    
      else {
        db.query UserCliConf {
          where = $db.UserCliConf.idCliente == $input.idCliente
          sort = {UserCliConf.nome: "asc"}
          return = {type: "list"}
        } as $UserCliConf1
      
        var.update $x1 {
          value = $UserCliConf1
        }
      }
    }
  }

  response = $x1
}