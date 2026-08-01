// Ativa/renova vis_licenca apos pagamento de fatura
function fn_fp_confvision_ativar_licenca_pagamento {
  input {
    int vis_licenca_id? filters=min:1
    int fp_fatura_id? filters=min:1
    int fp_pagamento_id? filters=min:1
    timestamp? pago_em?
  }

  stack {
    db.get vis_licenca {
      field_name = "id"
      field_value = $input.vis_licenca_id
    } as $lic
  
    precondition ($lic != null) {
      error = "Licenca nao encontrada"
    }

    conditional {
      if ($lic.unidade == "gravacao") {
        function.run fn_vis_gravacao_storage_ensure {
          input = {id_franqueado: $lic.id_franqueado}
        } as $storage_ensure
      }
    }
  
    conditional {
      if ($lic.unidade == "gravacao") {
        function.run fn_vis_gravacao_storage_ensure {
          input = {id_franqueado: $lic.id_franqueado}
        } as $storage_ensure
      }
    }
  
    var $pago {
      value = $input.pago_em|first_notempty:now
    }
  
    var $base {
      value = $lic.valido_ate
    }
  
    conditional {
      if (($base == null) || ($base < now)) {
        var.update $base {
          value = now
        }
      }
    }
  
    var $novo_valido {
      value = $base|add_secs_to_timestamp:30 * 86400
    }
  
    var $novo_status {
      value = $lic.status
    }
  
    conditional {
      if (($lic.status == "pendente") || ($lic.status == "expirada")) {
        var.update $novo_status {
          value = "disponivel"
        }
      }
    
      elseif (($lic.status == "disponivel") || ($lic.status == "em_uso")) {
        var.update $novo_status {
          value = $lic.status
        }
      }
    }
  
    db.patch vis_licenca {
      field_name = "id"
      field_value = $lic.id
      data = {
        valido_ate  : $novo_valido
        status      : $novo_status
        pago_em     : $pago
        id_fatura   : $input.fp_fatura_id|to_text
        id_pagamento: $input.fp_pagamento_id|to_text
      }
    } as $lic_upd
  }

  response = $lic_upd
}