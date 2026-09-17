// src/context/CartContext.tsx
import { createContext, useContext, useState, type ReactNode } from 'react';

// 1. Definimos las interfaces (tipos de datos)
export interface Producto {
  id: number;
  nombre: string;
  precio: number;
  img: string;
}

export interface CartItem extends Producto {
  cantidad: number;
}

interface CartContextType {
  cart: CartItem[];
  addToCart: (producto: Producto) => void;
  removeFromCart: (id: number) => void;
  totalItems: number;
  totalPrice: number;
}

// 2. Creamos el contexto indicando que puede ser CartContextType o undefined
const CartContext = createContext<CartContextType | undefined>(undefined);

// 3. Hook personalizado con validación de tipo
export const useCart = () => {
  const context = useContext(CartContext);
  if (!context) {
    throw new Error("useCart debe ser usado dentro de un CartProvider");
  }
  return context;
};

// 4. Tipamos los props del Provider
interface CartProviderProps {
  children: ReactNode;
}

// 5. El Provider
export const CartProvider = ({ children }: CartProviderProps) => {
  // Le decimos a useState que este arreglo contendrá objetos de tipo CartItem
  const [cart, setCart] = useState<CartItem[]>([]);

  const addToCart = (producto: Producto) => {
    setCart((prevCart) => {
      const itemExists = prevCart.find((item) => item.id === producto.id);
      if (itemExists) {
        return prevCart.map((item) =>
          item.id === producto.id
            ? { ...item, cantidad: item.cantidad + 1 }
            : item
        );
      }
      return [...prevCart, { ...producto, cantidad: 1 }];
    });
  };

  const removeFromCart = (id: number) => {
    setCart((prevCart) => prevCart.filter((item) => item.id !== id));
  };

  const totalItems = cart.reduce((total, item) => total + item.cantidad, 0);
  const totalPrice = cart.reduce((total, item) => total + item.precio * item.cantidad, 0);

  return (
    <CartContext.Provider value={{ cart, addToCart, removeFromCart, totalItems, totalPrice }}>
      {children}
    </CartContext.Provider>
  );
};