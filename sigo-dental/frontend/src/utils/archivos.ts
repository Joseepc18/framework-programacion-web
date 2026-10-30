// Entrega al navegador un archivo ya descargado con la sesión (el Bearer no viaja en un enlace normal).

// Lo abre en una pestaña nueva (PDF e imágenes) o lo guarda con el nombre dado.
export function abrirBlob(blob: Blob, nombre: string, guardar = false) {
  const url = URL.createObjectURL(blob);
  if (guardar) {
    const enlace = document.createElement('a');
    enlace.href = url;
    enlace.download = nombre;
    enlace.click();
  } else {
    window.open(url, '_blank', 'noopener');
  }
  // El navegador necesita el objeto un momento para cargarlo antes de liberarlo.
  setTimeout(() => URL.revokeObjectURL(url), 60_000);
}

export function tamanoLegible(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
