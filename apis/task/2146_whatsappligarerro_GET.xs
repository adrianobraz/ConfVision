// query all task wheres padrao
query whatsappligarerro verb=GET {
  api_group = "task"

  input {
    bool? falha?
    bool? exec?
    int? tentativas?=0
    bool? atendido?
    text? idProcesso? filters=trim
    int? limit?=0
  }

  stack {
    var $per_page {
      value = 10000
    }
  
    conditional {
      if ($input.limit != null && $input.limit > 0) {
        var.update $per_page {
          value = $input.limit
        }
      }
    }
  
    db.query whatsappLigarErro {
      where = ($input.falha == null || $db.whatsappLigarErro.falha == $input.falha) && ($input.exec == null || $db.whatsappLigarErro.exec == $input.exec) && ($input.tentativas == null || $db.whatsappLigarErro.tentativas == $input.tentativas) && ($input.atendido == null || $db.whatsappLigarErro.atendido == $input.atendido) && ($input.idProcesso == null || $db.whatsappLigarErro.idProcesso == $input.idProcesso)
      sort = {
        whatsappLigarErro.tentativas: "asc"
        whatsappLigarErro.id        : "asc"
      }
    
      return = {
        type  : "list"
        paging: {page: 1, per_page: $per_page, metadata: false}
      }
    } as $resultado
  }

  response = {dados: $resultado}
}