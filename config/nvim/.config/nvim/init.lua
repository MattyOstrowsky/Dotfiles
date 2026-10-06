-- Minimal offline editor; no plugin bootstrap on servers.
vim.opt.number = true
vim.opt.expandtab = true
vim.opt.shiftwidth = 2
vim.opt.tabstop = 2
vim.opt.ignorecase = true
vim.opt.smartcase = true
vim.opt.termguicolors = true
vim.opt.background = "dark"
vim.opt.undofile = true
vim.opt.signcolumn = "yes"
vim.cmd("syntax enable")
local groups = {
  Normal = { fg = "#d8dee9", bg = "#2e3440" },
  NormalFloat = { fg = "#d8dee9", bg = "#3b4252" },
  Comment = { fg = "#616e88", italic = true },
  Constant = { fg = "#b48ead" },
  String = { fg = "#a3be8c" },
  Identifier = { fg = "#d8dee9" },
  Function = { fg = "#88c0d0" },
  Statement = { fg = "#81a1c1" },
  PreProc = { fg = "#81a1c1" },
  Type = { fg = "#8fbcbb" },
  Special = { fg = "#ebcb8b" },
  Error = { fg = "#bf616a" },
  Visual = { bg = "#434c5e" },
  Search = { fg = "#2e3440", bg = "#ebcb8b" },
  LineNr = { fg = "#4c566a" },
  CursorLineNr = { fg = "#ebcb8b" },
  StatusLine = { fg = "#d8dee9", bg = "#3b4252" },
  Pmenu = { fg = "#d8dee9", bg = "#3b4252" },
  PmenuSel = { fg = "#eceff4", bg = "#434c5e" },
}
for group, value in pairs(groups) do vim.api.nvim_set_hl(0, group, value) end
vim.g.colors_name = "nord"
