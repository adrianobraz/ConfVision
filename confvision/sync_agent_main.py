"""Processo dedicado de sync — 1 instancia por VPS alimenta Redis para todos os workers."""

from sync_agent import agent_loop


if __name__ == "__main__":
    agent_loop()
