// Monta ciclo_ref ASS-{id}-{Ymd} para fatura de assinatura
function fn_fp_assinatura_ciclo_ref {
  input {
    int assinatura_id? filters=min:1
    timestamp? data_cobranca?
  }

  stack {
    var $id_txt {
      value = $input.assinatura_id|to_text
    }
  
    var $ymd {
      value = $input.data_cobranca|format_timestamp:"Ymd":"UTC"
    }
  
    var $ref {
      value = "ASS-" ~ $id_txt ~ "-" ~ $ymd
    }
  }

  response = $ref
}