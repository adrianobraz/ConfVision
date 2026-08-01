task Whatsapp_CoolDown {
  active = false

  stack {
    function.run fn_ClaimWhatsappProcFila as $func1
  }

  schedule = [{starts_on: 2026-03-28 16:00:00+0000, freq: 86400}]
}