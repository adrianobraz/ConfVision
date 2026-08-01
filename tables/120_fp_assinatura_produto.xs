table fp_assinatura_produto {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text id_franqueado? filters=trim
    text produto? filters=trim
    text plano? filters=trim
    text status?=pendente filters=trim
    text periodicidade?=mensal filters=trim
    decimal valor?
    decimal valor_piso_central?
    decimal valor_piso_breakglass?
    decimal margem_central?
    decimal margem_rep?
    text id_representante? filters=trim
    text id_central? filters=trim
    timestamp? valido_ate?
    timestamp? proxima_cobranca_em?
    timestamp? ultima_cobranca_em?
    json limites_json?
    json modulos_json?
    int fp_produto_catalogo_id? {
      table = "fp_produto_catalogo"
    }
  
    text ciclo_fatura_ref? filters=trim
    text observacao? filters=trim
    text tipo_contratacao?=pacote filters=trim
    json addons_json?
    json addons_pendentes_json?
    // Pacotes de cota comprados (lotes): [{id,nome,quantidade,valor}]
    json cotas_json?
    decimal credito_saldo?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "id_franqueado", op: "asc"}]}
    {type: "btree", field: [{name: "valido_ate", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "proxima_cobranca_em", op: "asc"}]
    }
    {
      type : "btree"
      field: [
        {name: "id_franqueado", op: "asc"}
        {name: "produto", op: "asc"}
        {name: "status", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [{name: "ciclo_fatura_ref", op: "asc"}]
    }
    {type: "btree", field: [{name: "id_representante", op: "asc"}]}
    {type: "btree", field: [{name: "id_central", op: "asc"}]}
  ]
}
