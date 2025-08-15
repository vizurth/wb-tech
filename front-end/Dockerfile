FROM nginx:alpine

# Копируем HTML и JS в стандартную папку Nginx
COPY . /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
