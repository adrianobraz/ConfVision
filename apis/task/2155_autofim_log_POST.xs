query "autofim/log" verb=POST {
  api_group = "task"

  input {
    int idEvento
    text idProcesso
    text idDispositivo
    text motivo
    text acao
    int qtdCiclos3mDiffProc?
    bool temFalhas?
    bool temAlarme?
    bool temDesarme?
    bool temRestaure?
    bool temParAlarmeRest50?
    json retornoProcessoEnd?
    text erro?
    text regraVersao?
  }

  stack {
    precondition (($input.idEvento|is_empty) == false) {
      error = "idEvento obrigatório"
      payload = false
    }
  
    precondition (($input.idProcesso|trim|is_empty) == false) {
      error = "idProcesso obrigatório"
      payload = false
    }
  
    precondition (($input.idDispositivo|trim|is_empty) == false) {
      error = "idDispositivo obrigatório"
      payload = false
    }
  
    precondition (($input.motivo|trim|is_empty) == false) {
      error = "motivo obrigatório"
      payload = false
    }
  
    precondition (($input.acao|trim|is_empty) == false) {
      error = "acao obrigatório"
      payload = false
    }
  
    var $qtdCiclos {
      value = 0
    }
  
    conditional {
      if (($input.qtdCiclos3mDiffProc|is_empty) == false) {
        var.update $qtdCiclos {
          value = $input.qtdCiclos3mDiffProc
        }
      }
    }
  
    var $versao {
      value = "vFinal_go_domain_api"
    }
  
    conditional {
      if (($input.regraVersao|trim|is_empty) == false) {
        var.update $versao {
          value = $input.regraVersao
        }
      }
    }
  
    db.add alarmEvent_autofimlog {
      enforce_hidden_fields = false
      data = {
        created_at        : "now"
        idEvento          : $input.idEvento
        idProcesso        : $input.idProcesso
        idDispositivo     : $input.idDispositivo
        motivo            : $input.motivo
        acao              : $input.acao
        qtdCiclos5m       : $qtdCiclos
        bloqueio3x5       : false
        temFalhas         : $input.temFalhas
        temAlarme         : $input.temAlarme
        temDesarme        : $input.temDesarme
        temRestaure       : $input.temRestaure
        temParAlarmeRest50: $input.temParAlarmeRest50
        retornoProcessoEnd: $input.retornoProcessoEnd
        regraVersao       : $versao
        erro              : $input.erro
      }
    } as $log1
  }

  response = {dados: ""|set:"ok":true|set:"idLog":$log1.id}
}