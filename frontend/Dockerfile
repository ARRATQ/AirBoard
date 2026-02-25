# Build stage
FROM node:18-alpine AS builder

# Définition du répertoire de travail
WORKDIR /app

# Copie des fichiers de dépendances
COPY package*.json ./

# Installation des dépendances (avec devDependencies pour le build)
RUN npm ci --silent

# Copie du code source
COPY . .

# Copie du fichier .env.production pour la configuration de production
COPY .env.production .env

# Build de l'application avec NODE_ENV=production
ENV NODE_ENV=production
RUN npm run build

# Production stage
FROM nginx:alpine

# Copie de la configuration Nginx
COPY nginx.conf /etc/nginx/nginx.conf

# Copie des fichiers buildés
COPY --from=builder /app/dist /usr/share/nginx/html

# Création d'un utilisateur non-root
RUN addgroup -g 1001 -S airboard && \
    adduser -S airboard -u 1001

# Ajustement des permissions pour fonctionner avec un utilisateur non-root
RUN chown -R airboard:airboard /usr/share/nginx/html && \
    chown -R airboard:airboard /var/cache/nginx && \
    chown -R airboard:airboard /var/log/nginx && \
    chown -R airboard:airboard /etc/nginx/conf.d && \
    # Créer les répertoires nécessaires avec les bonnes permissions
    mkdir -p /var/run/nginx /var/lib/nginx && \
    chown -R airboard:airboard /var/run/nginx /var/lib/nginx && \
    # Rendre nginx.conf accessible
    chown airboard:airboard /etc/nginx/nginx.conf

# Utilisation de l'utilisateur non-root
USER airboard

# Exposition du port
EXPOSE 80

# Variables d'environnement
ENV NODE_ENV=production

# Commande de démarrage
CMD ["nginx", "-g", "daemon off;"]