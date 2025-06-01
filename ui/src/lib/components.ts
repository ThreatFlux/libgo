import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "./utils";

// Button component
export const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        destructive:
          "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        outline:
          "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-9 rounded-md px-3",
        lg: "h-11 rounded-md px-8",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

// Simple Card components without JSX
export const Card = React.forwardRef(
  ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>, ref: React.Ref<HTMLDivElement>) => {
    const element = React.createElement("div", {
      ref,
      className: cn("rounded-lg border bg-card text-card-foreground shadow-sm", className),
      ...props
    });
    return element;
  }
);
Card.displayName = "Card";

export const CardHeader = React.forwardRef(
  ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>, ref: React.Ref<HTMLDivElement>) => {
    const element = React.createElement("div", {
      ref,
      className: cn("flex flex-col space-y-1.5 p-6", className),
      ...props
    });
    return element;
  }
);
CardHeader.displayName = "CardHeader";

export const CardTitle = React.forwardRef(
  ({ className, ...props }: React.HTMLAttributes<HTMLHeadingElement>, ref: React.Ref<HTMLHeadingElement>) => {
    const element = React.createElement("h3", {
      ref,
      className: cn("text-2xl font-semibold leading-none tracking-tight", className),
      ...props
    });
    return element;
  }
);
CardTitle.displayName = "CardTitle";

export const CardDescription = React.forwardRef(
  ({ className, ...props }: React.HTMLAttributes<HTMLParagraphElement>, ref: React.Ref<HTMLParagraphElement>) => {
    const element = React.createElement("p", {
      ref,
      className: cn("text-sm text-muted-foreground", className),
      ...props
    });
    return element;
  }
);
CardDescription.displayName = "CardDescription";

export const CardContent = React.forwardRef(
  ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>, ref: React.Ref<HTMLDivElement>) => {
    const element = React.createElement("div", {
      ref,
      className: cn("p-6 pt-0", className),
      ...props
    });
    return element;
  }
);
CardContent.displayName = "CardContent";

export const CardFooter = React.forwardRef(
  ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>, ref: React.Ref<HTMLDivElement>) => {
    const element = React.createElement("div", {
      ref,
      className: cn("flex items-center p-6 pt-0", className),
      ...props
    });
    return element;
  }
);
CardFooter.displayName = "CardFooter";

// Tabs Components
export const Tabs = React.forwardRef(
  ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>, ref: React.Ref<HTMLDivElement>) => {
    return React.createElement("div", {
      ref,
      className: cn("data-[state=active]:bg-background", className),
      ...props
    });
  }
);
Tabs.displayName = "Tabs";

export const TabsList = React.forwardRef(
  ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>, ref: React.Ref<HTMLDivElement>) => {
    return React.createElement("div", {
      ref,
      className: cn(
        "inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground",
        className
      ),
      ...props
    });
  }
);
TabsList.displayName = "TabsList";

export const TabsTrigger = React.forwardRef(
  ({ className, ...props }: React.HTMLAttributes<HTMLButtonElement>, ref: React.Ref<HTMLButtonElement>) => {
    return React.createElement("button", {
      ref,
      className: cn(
        "inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm",
        className
      ),
      ...props
    });
  }
);
TabsTrigger.displayName = "TabsTrigger";

export const TabsContent = React.forwardRef(
  ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>, ref: React.Ref<HTMLDivElement>) => {
    return React.createElement("div", {
      ref,
      className: cn(
        "mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        className
      ),
      ...props
    });
  }
);
TabsContent.displayName = "TabsContent";