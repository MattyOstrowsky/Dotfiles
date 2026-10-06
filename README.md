# Dotfiles

Konfiguracje + jeden instalator w Go/Bubble Tea. TUI i CLI korzystają z tego
samego mechanizmu. Domyślny motyw: **Nord Dark**. Bez GNU Stow, instalowania
pluginów przy starcie shella i automatycznego aktualizowania repozytorium.

```text
config/               konfiguracje aplikacji w układzie katalogu domowego
installer/            kod, katalog narzędzi, testy i przykłady wdrożeń
  tools.yaml          dostępne narzędzia i źródła instalacji
  examples/           profile YAML/JSON i playbook Ansible
Makefile
```

## Uruchomienie

Do kompilacji potrzebny jest Go 1.26+. Na Ubuntu: `sudo apt install golang-go make`.
Na docelowej VM wystarczy skompilowany plik, `config/` i wybrany profil.

```bash
make install                      # kompilacja i TUI
make list                         # katalog narzędzi
make test                         # testy Go i fish/Starship, jeśli dostępne
```

W TUI nic nie jest domyślnie zaznaczone. Spacja na nagłówku zaznacza/odznacza
kategorię; spacja na narzędziu zmienia jeden wybór. `↑/↓` lub `j/k` przewijają listę.

| Klawisz | Działanie |
|---|---|
| `a` | wszystkie / żadne |
| `i` | instalacja pakietów i binarek |
| `c` | wdrożenie konfiguracji wybranych aplikacji |
| `m` | kopiowanie plików / dowiązania |
| `b` | kopia zapasowa przy konflikcie / błąd |
| `s` | zapis wyboru do `dotfiles-selection.yaml` w bieżącym katalogu |
| `Enter` | podsumowanie, potem wykonanie |
| `q` | wyjście bez instalacji |

Zapis nie nadpisuje istniejącego profilu. Można wczytać profil do TUI:

```bash
./installer/dotfiles-install --profile installer/examples/vps.yaml --tui
```

Kategorie: `general`, `shell`, `git`, `containers`, `monitoring`, `cloud`,
`kubernetes`, `automation`, `development`, `editors`, `ai`.
Cloud i Kubernetes są opcjonalne; profil VPS ich nie zawiera.

## Profil i CLI

`workstation.yaml` odwzorowuje narzędzia obecne na lokalnym hoście. Ma
`install: false`, `configs: true` i `config_mode: link`: wyrównuje konfiguracje
z repo, korzystając z istniejących binarek, także Docker Desktop i lokalnego
Neovima. Zastosowanie: `./installer/dotfiles-install --profile
installer/examples/workstation.yaml --yes` (polecenie w jednym wierszu).

```yaml
version: 1
tools: [fish, starship, git, docker, bat, btop]
# categories: [general, git]
# exclude: [wget]
# packages: [ca-certificates, less]
install: true
configs: true
config_mode: copy
conflict: backup
```

`tools` i `categories` sumują się; `exclude` usuwa pozycje z wyniku.
`packages` to dodatkowe **natywne nazwy pakietów** danej dystrybucji, bez
argumentów powłoki. Nie otrzymują automatycznie konfiguracji. YAML i JSON mają
ten sam schemat. Nieznane pola/narzędzia/kategorie kończą się błędem.

```bash
# Tylko podgląd: bez sudo, sieci, odświeżania APT ani zapisów.
./installer/dotfiles-install --profile installer/examples/vps.yaml --dry-run

# Wdrożenie bez interakcji; logi na stderr, wynik JSON na stdout.
./installer/dotfiles-install --profile installer/examples/vps.yaml --yes --json

# Wybór bez pliku.
./installer/dotfiles-install --tools fish,starship,git --configs-only --yes
./installer/dotfiles-install --categories general,git --no-configs --dry-run
./installer/dotfiles-install --tools git --packages jq,less --no-configs --yes

# Tylko konfiguracje, z JSON.
./installer/dotfiles-install --profile installer/examples/config-only.json --yes
```

Flagi wyboru `--tools`, `--categories`, `--exclude`, `--packages` zastępują
odpowiadające im listy z pliku. `--no-configs` wyłącza konfiguracje;
`--configs-only` wyłącza instalację i włącza konfiguracje.
`--config-mode copy|link` i `--conflict backup|error` nadpisują profil.
`--yes` wymaga jawnego wyboru — samo `--yes` nie instaluje wszystkiego.

Przykładowy wynik:

```json
{"changed":false,"dry_run":false,"actions":[{"kind":"config","name":"/home/ubuntu/.config/fish/config.fish","status":"ok"}]}
```

Kod wyjścia: `0` sukces, niezerowy błąd. Przy błędzie JSON zawiera `error`.
Przy `--dry-run` pole `changed` oznacza przewidywane zmiany; dostępność paczek
na serwerze repozytorium i pobieranie wydań są sprawdzane podczas wykonania.
Ponowne wykonanie pomija zgodne pliki i zainstalowane pakiety/binarki.
Instalator zapewnia obecność narzędzi, **nie aktualizuje ich do najnowszych wersji**.

## Pliki, uprawnienia i kopie zapasowe

`--repo` wskazuje katalog zawierający `config/`; domyślnie szukamy obok binarki
i w bieżącym katalogu. `--target` wskazuje katalog domowy (domyślnie `$HOME`).
Konfiguracje wdrażaj jako ich właściciel. Pakiety systemowe wymagają roota lub
sudo; CLI używa `sudo -n`, więc nie zatrzyma Ansible na pytaniu o hasło.
W TUI sudo może poprosić o hasło po opuszczeniu ekranu wyboru.

Binarki upstream trafiają do `TARGET/.local/bin`; `--bin-dir /usr/local/bin`
umieszcza je wspólnie dla wszystkich użytkowników. Ubuntu otrzymuje również
brakujące skróty binarek `bat → /usr/bin/batcat` i `fd → /usr/bin/fdfind`.
To osobne pliki wykonywalne; `cat` pozostaje zwykłym `cat`.

Domyślne `copy` jest odpowiednie dla Ansible: docelowy config nie zależy od
obecności checkoutu. `link` pozwala edytować konfigurację bezpośrednio w repo.
Zmieniane pliki/dowiązania są najpierw przenoszone obok, do unikalnych
`.dotfiles-backup-NAZWA-*`. Ich ścieżki są w wyniku. Aby przywrócić konfigurację,
przenieś wybraną kopię pod pierwotną nazwę po usunięciu wdrożonego pliku.

Stare katalogowe dowiązania Stow do tego repo są przy `copy` zamieniane na
zwykłe katalogi, z zachowaniem plików lokalnych. Dowiązania katalogów prowadzące
poza wybrany pakiet konfiguracji powodują błąd. `conflict: error` przerywa
przed instalacją pakietów, jeżeli konfiguracje kolidują.
Instalator nie usuwa niezarządzanych plików z katalogu docelowego.

APT, DNF i pacman korzystają z repozytoriów skonfigurowanych na hoście.
APT używa `--no-install-recommends`, DNF wyłącza słabe zależności.
Pacman nie wykonuje osobnego `-Sy` ani automatycznej aktualizacji całego systemu;
wcześniej utrzymuj system przez normalne `pacman -Syu`.
Brak mapowania pakietu dla danej dystrybucji powoduje jawny błąd.
Wydania upstream dla amd64/arm64 są weryfikowane przez SHA-256; brak sumy lub
niezgodność przerywa instalację. Nie ma `curl | bash`.
`terraform`, `opencode`, `omp` są jawnie oznaczone jako zewnętrzne:
instalator obsługuje ich dostępne konfiguracje, a nie instalację programów.

## Fish, prompt i Nord

Fish zachowuje odziedziczony PATH, usuwa dokładne duplikaty i dodaje istniejące
katalogi `~/.local/bin`, Cargo, Bun, fnm, Atuin i Go tylko raz. Nie zapisuje
`fish_user_paths`; respektuje `GOPATH`, `CARGO_HOME`, `BUN_INSTALL` i `EDITOR`.
Ścieżki z aktywnego venv oraz WSL pozostają. Dodatki startują tylko, gdy są
zainstalowane. `fish_variables` jest lokalnym stanem, nie częścią wdrożenia.

Skróty są jawnymi abbreviations (`gs`, `gd`, `dc`, `ll`). Konfiguracja nie
podmienia `cat`, `bat`, `grep`, `top`, `ls` ani `cd`. Starship używa dwuwierszowego układu
`┌─>` / `└─>` z czasem, Gitem i kontekstowymi modułami środowiska. Po zmianach uruchom nowy terminal
lub `exec fish`, aby pozbyć się funkcji utworzonych przez stary config.

Nord obejmuje fish/fzf, Starship, Git diff, bat, btop, lazygit, lazydocker,
k9s, glow, Atuin, tmux, Neovim, OpenCode i OMP. Neovim ma mały config bez pluginów,
klonowania repozytoriów czy zależności od Nerd Fonts. Git nie ustawia nazwiska,
maila ani credential helpera — lokalna tożsamość pozostaje w `~/.gitconfig`.
Docker i zwykłe CLI bez obsługi motywów używają kolorów terminala.

## Ansible / małe VPS-y Ubuntu 26.04

Gotowy przykład: `installer/examples/ansible/deploy.yml`. Utwórz własne inventory
na podstawie `inventory.example.yml`. Użytkownik docelowy musi już istnieć.
Na kontrolerze zbuduj binarkę dla architektury VM:

```bash
make build
# Dla ARM: cd installer && GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o dotfiles-install .
ansible-playbook -i installer/examples/ansible/inventory.local.yml \
  installer/examples/ansible/deploy.yml --ask-become-pass
```

Playbook kopiuje binarkę, konfiguracje i profil. Pakiety instaluje jako root,
wydania upstream umieszcza w `/usr/local/bin`, a konfiguracje kopiuje jako
użytkownik docelowy. `changed_when` czyta `changed` z JSON. Profile z
`install: false` lub `configs: false` wyłączają odpowiedni etap.
Można nadpisać `dotfiles_profile`, `dotfiles_source`, `dotfiles_remote` i `dotfiles_user`.

`--check` uruchamia instalator z `--dry-run` na już wdrożonych plikach.
Na zupełnie nowym hoście pokazuje plan kopiowania, a wykonanie binarki pomija.
Zmiany lokalnego profilu/configów trzeba najpierw wdrożyć, aby plan instalatora
odpowiadał ich nowej wersji. Go i Ansible nie są potrzebne na VM;
Ansible wymaga tam Python 3 i dostępu przez SSH.

Profil `vps.yaml` jest mały: fish, Git, Docker/Compose/Buildx,
wyszukiwanie, jq, bat, btop i tmux. Nie instaluje Starshipa: na świeżym VPS
fish używa swojego zwykłego promptu. Wspólny config fish uruchamia Starshipa
tylko tam, gdzie jego binarka jest zainstalowana. Docker pochodzi z repozytorium Ubuntu
(`docker.io`, `docker-compose-v2`, `docker-buildx`), nie z mieszaniny pakietów
Ubuntu i Docker CE. Na hoście z Docker CE/Desktop wybierz `exclude: [docker]`.

Dla serwera ustaw niezależnie od dotfiles:

- Dostęp SSH kluczem dla zwykłego użytkownika z sudo; zmiany uwierzytelniania
  wdrażaj po sprawdzeniu nowej sesji SSH.
- Włączone aktualizacje bezpieczeństwa i zaplanowane restarty. Ubuntu opisuje
  konfigurację [unattended-upgrades](https://documentation.ubuntu.com/server/how-to/software/automatic-updates/).
- Limit logów kontenerów, np. w Compose `logging: {driver: local}`; driver
  [local](https://docs.docker.com/engine/logging/drivers/local/) ma rotację.
- Wystawiaj publicznie tylko potrzebne porty; bazy i usługi wewnętrzne binduj
  do `127.0.0.1`. Opublikowane porty Dockera mogą omijać UFW — opisuje to
  [dokumentacja Dockera](https://docs.docker.com/engine/network/packet-filtering-firewalls/).
- Kopie danych poza VM i test odtwarzania; btop służy do podglądu, nie zastępuje
  monitoringu dostępności i wolnego miejsca.

Konfiguracja systemowego SSH, firewalla, domyślnego shella i grupy `docker`
pozostaje w playbooku zarządzającym hostem, aby profil narzędzi nie zmieniał
przypadkiem dostępu administracyjnego.
