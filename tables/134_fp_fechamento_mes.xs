table fp_fechamento_mes {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text competencia? filters=trim
    text status?=fechado filters=trim
    timestamp? fechado_em?
    text fechado_por? filters=trim
    decimal saldo_caixa_inicio?
    decimal saldo_caixa_fim?
    decimal total_entradas?
    decimal total_saidas?
    decimal total_receitas?
    decimal total_despesas?
    decimal total_a_receber?
    decimal total_vencido?
    int qtd_inadimplentes?
    int assinaturas_ativas?
    int assinaturas_pendentes?
    int faturas_pagas?
    int faturas_abertas?
    text snapshot_json? filters=trim
    text observacao? filters=trim
    text id_central? filters=trim
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "competencia", op: "asc"}]}
    {type: "btree", field: [{name: "status", op: "asc"}]}
    {type: "btree", field: [{name: "id_central", op: "asc"}]}
  ]
}