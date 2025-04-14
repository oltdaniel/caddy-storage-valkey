#!/usr/bin/env bash

# Output file
output_file="Caddyfile"
num=5000

# Clear the file if it exists and write the initial config
cat > "$output_file" <<EOF
{
    debug

    storage valkey {
        address localhost:6379
    }

    storage_clean_interval 60s
}
EOF

# Function to generate a random domain
gen_domain() {
    tlds=("com" "net" "org" "io" "xyz" "dev" "app" "tech" "site")
    tld=${tlds[$RANDOM % ${#tlds[@]}]}
    echo "$(uuidgen).$tld"
}

# Generate entries
for ((i = 1; i <= $num; i++)); do
    domain=$(gen_domain)
    cat >> "$output_file" <<EOF
$domain {
    tls {
        issuer internal {
            lifetime 1h
        }
    }
    respond "Hello World" 200
}
EOF
done

echo "Generated $num random domain entries in $output_file"
