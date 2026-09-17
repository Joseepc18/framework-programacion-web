# GUÍA DE PRÁCTICA DE LABORATORIO

**Asignatura:** Framework Programación Web[cite: 3]
**Unidad 1:** Frameworks Frontend para Plataformas Interactivas[cite: 3]
**Tema:** Configuración de entornos SPA modernos con Vite y Tailwind CSS[cite: 3]

---

### **Objetivo de la Práctica**

Transicionar de la teoría arquitectónica a la ejecución práctica mediante el levantamiento de un entorno de desarrollo para una **SPA** (_Single Page Application_). Se comprobará empíricamente la velocidad de arranque del empaquetador **Vite**, el funcionamiento del **HMR** (_Hot Module Replacement_) y la inyección de estilos utilitarios de la arquitectura moderna de **Tailwind CSS v4** (CSS-first).

### **Requisitos Previos**

- Entorno de ejecución **Node.js** instalado.
- Un IDE moderno (Ej. Visual Studio Code).
- Terminal de comandos abierta.

---

### **Fase 1: Inicialización del Empaquetador Ágil (Vite)**

Reemplazaremos los _bundlers_ tradicionales para aprovechar el **ESM** (_ECMAScript Modules_) nativo del navegador, garantizando un arranque instantáneo.

1. Abre tu terminal y ejecuta el comando de andamiaje (_scaffolding_) para crear el proyecto base:

   npm create vite@latest mi-proyecto-spa

El asistente interactivo solicitará configurar el entorno. Selecciona las siguientes opciones:

Framework: React (o Vue, según el enfoque definido).

Variant: JavaScript o TypeScript.

Which linter to use? Selecciona ESLint (esta herramienta garantizará que el código mantenga un formato estandarizado y profesional).

Navega al directorio recién creado e instala las dependencias base:

    cd mi-proyecto-spa
    npm install

Fase 2: Integración de Tailwind CSS (Estándar v4)
Implementaremos la versión moderna de Tailwind CSS, la cual elimina la necesidad de archivos de configuración antiguos (tailwind.config.js y postcss.config.js) y funciona directamente como un plugin de Vite.

1. Instala el framework base y su plugin oficial para Vite:

   npm install tailwindcss @tailwindcss/vite

2. Abre el archivo vite.config.js en la raíz de tu proyecto e importa el plugin de Tailwind para orquestarlo junto a React/Vue:

   import { defineConfig } from 'vite'
   import react from '@vitejs/plugin-react'
   import tailwindcss from '@tailwindcss/vite'

   export default defineConfig({
   plugins: [
   react(),
   tailwindcss(),
   ],
   })

3. Abre el archivo principal de estilos (usualmente src/index.css o src/style.css), borra todo su contenido y añade la directiva de importación única del framework:

   @import "tailwindcss";

Fase 3: Ejecución y Comprobación de Sinergia (HMR + JIT)
Levantaremos el servidor para comprobar la inyección de código en vivo sin recargas y la compilación de estilos in-line.

1. Inicia el servidor de desarrollo:

   npm run dev

   (Observa el tiempo de arranque en la terminal; debe ser de unos pocos milisegundos).

2. Abre en tu navegador la ruta local indicada (usualmente http://localhost:5173).

3. Prueba de HMR y Estilos Utilitarios: Abre el archivo principal de tu aplicación (Ej. src/App.jsx o src/App.vue), borra el contenido por defecto e inyecta la siguiente estructura:

   function App() {
   return (
   <div className="flex h-screen w-full items-center justify-center bg-slate-900">
   <div className="text-center">
   <h1 className="text-4xl font-bold text-white mb-4">
   Entorno SPA Configurado
   </h1>
   <button className="bg-blue-500 hover:bg-blue-600 text-white font-semibold py-2 px-6 rounded-lg shadow-lg transition-all">
   Botón Utilitario
   </button>
   </div>
   </div>
   )
   }

   export default App

4. Validación: Guarda el archivo. La interfaz en el navegador se actualizará instantáneamente aplicando el diseño completo. Esto comprueba que el HMR de Vite y el compilador de Tailwind están operando correctamente en memoria.

Fase 4: Análisis de la Arquitectura de Archivos
Compara tu explorador de archivos para entender la función de cada nodo crítico en este ecosistema:

index.html: Punto de entrada único donde se inyecta el script principal, rasgo definitorio de una aplicación SPA.

src/: Directorio central que contiene los componentes de interfaz, hojas de estilo base y la lógica de negocio.

vite.config.js: Archivo orquestador que fusiona el servidor de desarrollo, el framework frontend (React/Vue) y el motor de estilos (Tailwind v4).
