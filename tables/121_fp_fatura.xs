table fp_fatura {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text id_franqueado? filters=trim
    text id_representante? filters=trim
    text id_central? filters=trim
    text referencia? filters=trim
    text status?=aberta filters=trim
    // assinatura | repasse_rep_central | repasse_central_breakglass | ...
    text tipo?=assinatura filters=trim
    decimal valor_total?
    decimal valor_piso_central?
    decimal valor_piso_breakglass?
    decimal margem_central?
    decimal margem_rep?
    int fatura_origem_id? {
      table = "fp_fatura"
    }
  
    timestamp? vencimento_em?
    timestamp? pago_em?
    text ciclo_ref? filters=trim
    text observacao? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "id_franqueado", op: "asc"}]}
    {type: "btree", field: [{name: "id_representante", op: "asc"}]}
    {type: "btree", field: [{name: "id_central", op: "asc"}]}
    {type: "btree", field: [{name: "status", op: "asc"}]}
    {type: "btree", field: [{name: "tipo", op: "asc"}]}
    {type: "btree", field: [{name: "vencimento_em", op: "asc"}]}
    {type: "btree", field: [{name: "referencia", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "id_franqueado", op: "asc"}
        {name: "ciclo_ref", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [{name: "fatura_origem_id", op: "asc"}]
    }
    {type: "btree", field: [{name: "id_central", op: "asc"}]}
  ]
}
