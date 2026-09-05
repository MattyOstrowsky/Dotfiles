require "nvchad.mappings"

-- add yours here

local map = vim.keymap.set

map("n", ";", ":", { desc = "CMD enter command mode" })
map("i", "jk", "<ESC>")

-- map({ "n", "i", "v" }, "<C-s>", "<cmd> w <cr>")

-- toggle render-markdown
map("n", "<leader>md", "<cmd>RenderMarkdown toggle<CR>", { desc = "Toggle markdown rendering" })

-- Trouble diagnostics list
map("n", "<leader>xx", "<cmd>Trouble diagnostics toggle<CR>", { desc = "Trouble diagnostics" })
map("n", "<leader>xw", "<cmd>Trouble diagnostics toggle filter.buf=0<CR>", { desc = "Trouble buffer diagnostics" })

-- Undotree
map("n", "<leader>u", "<cmd>UndotreeToggle<CR>", { desc = "Toggle undo tree" })

-- Todo comments
map("n", "<leader>td", "<cmd>TodoTrouble<CR>", { desc = "Todo comments" })
