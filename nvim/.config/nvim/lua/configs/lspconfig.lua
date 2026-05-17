require("nvchad.configs.lspconfig").defaults()

local servers = {
  "html", "cssls",
  "yamlls",
  "dockerls",
  "bashls",
  "jsonls",
  "terraformls",
  "ansiblels",
  "lua_ls",
}
vim.lsp.enable(servers)

-- read :h vim.lsp.config for changing options of lsp servers 
