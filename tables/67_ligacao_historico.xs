table LigacaoHistorico {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    text conversation_id? filters=trim
    bool success?
    text whatsapp_call_id? filters=trim
    int id_session?
    text caller? filters=trim
    text receiver? filters=trim
    text status? filters=trim
    text action? filters=trim
    text type? filters=trim
    text direction? filters=trim
    int id_user?
    int duration?
    text e_event_type? filters=trim
    timestamp? e_event_timestamp?
    text e_agent_id? filters=trim
    text e_agent_name? filters=trim
    text e_user_id? filters=trim
    text e_failure_reason? filters=trim
    int e_sip_status_code?
    text e_error_reason? filters=trim
    text e_twirp_code? filters=trim
    text e_sip_status? filters=trim
    int whatsappligarerro_id? {
      table = "whatsappLigarErro"
    }
  
    text call_sid? filters=trim
    text external_number? filters=trim
    text agent_number? filters=trim
    text termination_reason? filters=trim
    text analysis_success? filters=trim
    text analysis_summary_title? filters=trim
    text analysis_transcript_summary? filters=trim
    int cost_total?
    int call_charge_credits?
    int llm_charge_credits?
    decimal llm_price_usd?
    text tier? filters=trim
    bool is_burst?
    int free_minutes_consumed?
    decimal free_llm_dollars_consumed?
    int llm_input_tokens?
    int llm_output_tokens?
    decimal llm_input_price?
    decimal llm_output_price?
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {
      type : "btree"
      field: [{name: "conversation_id", op: "asc"}]
    }
    {
      type : "btree"
      field: [{name: "whatsapp_call_id", op: "asc"}]
    }
    {type: "btree", field: [{name: "receiver", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "whatsappligarerro_id", op: "asc"}]
    }
  ]
}