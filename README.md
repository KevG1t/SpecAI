# SpecAI

¡Bienvenido a **SpecAI**! 🚀

SpecAI es una herramienta de terminal (CLI) y una interfaz de usuario en consola (TUI) diseñada para gestionar, instalar y sincronizar reglas locales de Inteligencia Artificial y configuraciones de agentes en tus repositorios. Forma parte del ecosistema de herramientas de desarrollo inspirado en `gentle-ai`.

Si querés configurar rápidamente las reglas base de IA en tu proyecto para que Cursor, Windsurf o Codex entiendan tus estándares (o los del equipo), SpecAI se encarga de inyectar todo de manera automática y estructurada.

---

## 🛠️ Instalación

### Instalación Rápida (Recomendada)

La forma más sencilla de instalar SpecAI es usando nuestro script de instalación automática que detecta tu sistema operativo y arquitectura:

```bash
curl -fsSL https://kevg1t.github.io/SpecAI/install.sh | bash
```

Este script funciona en:
- 🐧 **Linux** (x86_64, ARM64, ARM, i386)
- 🍎 **macOS** (Intel y Apple Silicon)
- 🪟 **Windows** (WSL recomendado)

### Instalación Manual con Go

Si preferís compilar desde el código fuente:

```bash
go install github.com/KevG1t/SpecAI/cmd/specai@latest
```

*(Asegurate de tener tu `GOPATH` configurado en el `PATH` de tu sistema)*

### Verificar Instalación

Después de instalar, verificá que todo funcione:

```bash
specai --version
```

---

## 🚀 Cómo usar SpecAI

SpecAI tiene dos modos de funcionamiento principales. Posicionate en la carpeta del repositorio donde quieras inyectar las reglas y abrí la terminal.

### 1. Modo Interactivo (TUI)
Si ejecutás el comando sin ningún argumento adicional, se abrirá la interfaz gráfica en la misma consola. Podés navegar usando las flechas del teclado.
```bash
specai
```
¿Qué podés hacer desde acá?
- **Sync:** Sincronizar reglas y actualizar configuración.
- **Backup:** Administrar copias de seguridad de las configuraciones antes de modificarlas.
- **Uninstall:** Limpiar el repositorio de cualquier configuración previa de SpecAI.
- **Actualizaciones:** Revisar si tenés instalada la última versión disponible.

### 2. Modo CLI Directo (Instalación Rápida)
Si ya sabés lo que querés hacer y no tenés ganas de navegar por los menús, podés decirle a SpecAI que instale directamente el *setup* para tu editor de código.
```bash
# Instalar reglas de IA para Cursor
specai setup cursor

# Instalar reglas de IA para Windsurf
specai setup windsurf

# Instalar reglas de IA para Github Copilot / Codex
specai setup codex
```

---

## 🏗️ Arquitectura del Proyecto (Guía para Juniors)

Para los que recién arrancan a mirar el código, el repositorio está estructurado siguiendo los [estándares clásicos de Go](https://github.com/golang-standards/project-layout):

- `cmd/specai/main.go`: Es el **punto de entrada** del programa. Cuando ejecutás el binario, el código empieza a correr desde la función `main()` que está acá. Determina si abrimos la TUI o si corremos un comando directo (como `setup`).
- `internal/`: Todo el código de esta carpeta es privado para este proyecto (no se puede importar desde otros repositorios).
  - `internal/backup/`: Lógica para realizar "snapshots" (copias de seguridad) comprimidas del estado antes de hacer cambios destructivos.
  - `internal/catalog/`: Módulo que administra los datos y la lista de agentes o habilidades (*skills*) disponibles para instalar.
  - `internal/pipeline/`: ¡El corazón de la ejecución! Es un orquestador de tareas. Toma una lista de pasos, los ejecuta uno a uno, maneja los errores y, si algo sale mal a la mitad, se encarga del *rollback* (deshacer los cambios para no dejar cosas rotas).
  - `internal/steps/`: Contiene los "pasos individuales" (tareas aisladas) que utiliza el pipeline. Por ejemplo, `NewStepSetupLocalRules` es la cajita de código que contiene exclusivamente la lógica de inyectar las reglas.
  - `internal/system/`: Se encarga de hablar con el sistema operativo (manejar rutas, detectar qué IDE está instalado, escanear configuraciones).
  - `internal/tui/`: Todo lo relacionado con "dibujar" en la consola usando la librería BubbleTea (componentes, modelos, eventos de teclado).
  - `internal/update/`: Se conecta a la API de GitHub para revisar qué versión remota hay y avisarle al usuario si existe una versión más nueva.
  - `internal/verify/`: Contiene chequeos pre-vuelo (ej: asegurarse de que el entorno esté sano antes de intentar instalar algo).

---

## 🤝 Cómo Contribuir

Si te interesa mejorar el código, ¡metele mano!
1. Creá una rama nueva para tu desarrollo (`git checkout -b feature/mi-idea`).
2. Escribí tu magia respetando la arquitectura de la carpeta `internal/`.
3. Validá que no rompiste nada corriendo todos los tests automatizados:
   ```bash
   go test ./... -v
   ```
4. Confirmá tus cambios (`git commit -m "feat: agregué X cosa"`) y abrí un Pull Request en GitHub.
