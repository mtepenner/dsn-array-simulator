FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Copy proto files and generate Python stubs
COPY proto/ /app/proto_src/
RUN python -m grpc_tools.protoc \
    -I/app/proto_src \
    --python_out=/app/proto \
    --grpc_python_out=/app/proto \
    /app/proto_src/dsn.proto

RUN mkdir -p /app/proto && touch /app/proto/__init__.py

EXPOSE 50051

CMD ["python", "-m", "app.main"]
