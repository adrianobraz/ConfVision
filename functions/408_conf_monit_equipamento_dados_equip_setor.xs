function ConfMonitEquipamento_DadosEquipSetor {
  input {
    text Authorization? filters=trim
    text idDispositivo? filters=trim
    text particao? filters=trim
    text zonauser? filters=trim
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2000/v4/setor/listaByIdDispositivo"
      method = "POST"
      params = {}
        |set:"idDispositivo":$input.idDispositivo
      headers = []
        |push:"Content-Type: application/json"
        |push:("Authorization: Bearer"|concat:$input.Authorization:" ")
    } as $ap1
  
    conditional {
      if ((($var.ap1.response.result.status|is_empty)!=true) && ($var.ap1.response.result.status != "Vazio")) {
        var $x1 {
          value = $ap1.response.result.dados
            |lambda_filter:"""
            if (!$input.particao) return true;
            
            if (!$input.zonauser) {
              return $this.particao === $input.particao;
            }
            
            return $this.particao === $input.particao &&
                   $this.numero === $input.zonauser;
            """:10
        }
      }
    
      else {
        var $x1 {
          value = {}
        }
      }
    }
  }

  response = $x1
}