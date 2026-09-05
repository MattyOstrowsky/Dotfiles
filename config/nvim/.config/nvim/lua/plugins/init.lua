return {
  {
    "stevearc/conform.nvim",
    -- event = 'BufWritePre', -- uncomment for format on save
    opts = require "configs.conform",
  },

  -- These are some examples, uncomment them if you want to see them work!
  {
    "neovim/nvim-lspconfig",
    config = function()
      require "configs.lspconfig"
    end,
  },

  -- Markdown renderer — makes .md files readable in nvim
  {
    "MeanderingProgrammer/render-markdown.nvim",
    opts = {},
    ft = { "markdown" },
    cmd = "RenderMarkdown",
  },

  -- Treesitter — syntax highlighting
  {
    "nvim-treesitter/nvim-treesitter",
    opts = {
      ensure_installed = {
        "vim", "lua", "vimdoc",
        "markdown", "markdown_inline",
        "yaml", "dockerfile", "bash", "json", "toml", "hcl", "terraform",
      },
    },
  },

  -- Trouble — better diagnostics list
  {
    "folke/trouble.nvim",
    cmd = "Trouble",
    opts = {},
  },

  -- Todo Comments — highlight TODO/FIXME/HACK
  {
    "folke/todo-comments.nvim",
    cmd = "TodoTrouble",
    opts = {},
  },

  -- Undotree — visual undo history
  {
    "mbbill/undotree",
    cmd = "UndotreeToggle",
  },
}
